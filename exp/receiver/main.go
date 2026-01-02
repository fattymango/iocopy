package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	frame image.Image
	mu    sync.RWMutex
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
	addr := ":9000"

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer ln.Close()

	log.Println("[receiver] listening on", addr)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatal("accept error:", err)
	}
	defer conn.Close()

	log.Println("[receiver] client connected")

	for {
		var size uint32
		if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
			log.Println("read size error:", err)
			return
		}

		buf := make([]byte, size)
		if _, err := io.ReadFull(conn, buf); err != nil {
			log.Println("read frame error:", err)
			return
		}

		img, err := jpeg.Decode(bytes.NewReader(buf))
		if err != nil {
			log.Println("jpeg decode error:", err)
			continue
		}

		mu.Lock()
		frame = img
		mu.Unlock()

		log.Printf("[receiver] frame received: %d bytes\n", size)
	}
}

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	mu.RLock()
	img := frame
	mu.RUnlock()

	if img == nil {
		return
	}

	eimg := ebiten.NewImageFromImage(img)

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(eimg, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
