package main

import (
	"encoding/binary"
	"image"
	"io"
	"log"
	"net"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	frame    *image.RGBA
	mu       sync.RWMutex
	ebiImage *ebiten.Image

	width32  int32
	height32 int32
	width    int
	height   int
)

func main() {
	go startServer()

	ebiten.SetWindowSize(1280, 800)
	ebiten.SetWindowTitle("Screen Stream Viewer")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

func startServer() {
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer ln.Close()

	log.Println("[receiver] listening on :9000")

	conn, err := ln.Accept()
	if err != nil {
		log.Fatal("accept error:", err)
	}
	defer conn.Close()

	log.Println("[receiver] client connected")

	// TCP optimization
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetNoDelay(true)
	}

	// read width & height once
	if err := binary.Read(conn, binary.BigEndian, &width32); err != nil {
		log.Fatal(err)
	}
	if err := binary.Read(conn, binary.BigEndian, &height32); err != nil {
		log.Fatal(err)
	}

	width = int(width32)
	height = int(height32)

	log.Printf("[receiver] resolution: %dx%d\n", width, height)

	var buf []byte

	for {
		var size uint32
		if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
			log.Println("read size error:", err)
			return
		}

		// reuse buffer
		if cap(buf) < int(size) {
			buf = make([]byte, size)
		}
		buf = buf[:size]

		if _, err := io.ReadFull(conn, buf); err != nil {
			log.Println("read frame error:", err)
			return
		}

		img := &image.RGBA{
			Pix:    buf,
			Stride: 4 * width,
			Rect:   image.Rect(0, 0, width, height),
		}

		mu.Lock()
		frame = img
		mu.Unlock()
	}
}

type Game struct{}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebiten.Image) {
	mu.RLock()
	img := frame
	mu.RUnlock()

	if img == nil {
		return
	}

	// Create once
	if ebiImage == nil {
		ebiImage = ebiten.NewImageFromImage(img)
	} else {
		ebiImage.ReplacePixels(img.Pix)
	}

	screen.DrawImage(ebiImage, nil)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
