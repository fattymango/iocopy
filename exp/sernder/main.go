package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/kbinani/screenshot"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	log.Println("[sender] connected")

	// TCP optimization
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetNoDelay(true)
	}

	// get screen once
	bounds := screenshot.GetDisplayBounds(0)
	width := bounds.Dx()
	height := bounds.Dy()

	// send resolution once
	if err := binary.Write(conn, binary.BigEndian, int32(width)); err != nil {
		log.Fatal(err)
	}
	if err := binary.Write(conn, binary.BigEndian, int32(height)); err != nil {
		log.Fatal(err)
	}

	log.Printf("[sender] resolution: %dx%d\n", width, height)

	ticker := time.NewTicker(time.Second / 60) // ~60 FPS
	defer ticker.Stop()

	for range ticker.C {
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			log.Println("capture error:", err)
			continue
		}

		data := img.Pix

		// send size
		if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
			log.Println("write size error:", err)
			return
		}

		// send raw pixels
		_, err = conn.Write(data)
		if err != nil {
			log.Println("write frame error:", err)
			return
		}
	}
}
