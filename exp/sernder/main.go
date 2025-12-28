package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"log"
	"net"
	"time"

	"github.com/kbinani/screenshot"
)

func main() {
	addr := "localhost:9000" // RECEIVER IP
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	log.Println("[sender] connected to", addr)

	for {
		img, err := captureScreen()
		if err != nil {
			log.Println("capture error:", err)
			continue
		}

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 60})
		if err != nil {
			log.Println("jpeg error:", err)
			continue
		}

		data := buf.Bytes()

		// write frame size
		err = binary.Write(conn, binary.BigEndian, uint32(len(data)))
		if err != nil {
			log.Println("write size error:", err)
			return
		}

		// write frame
		_, err = conn.Write(data)
		if err != nil {
			log.Println("write frame error:", err)
			return
		}

		log.Printf("[sender] sent frame: %d bytes\n", len(data))
		time.Sleep(50 * time.Millisecond) // ~20 FPS
	}
}

func captureScreen() (*image.RGBA, error) {
	// capture primary display only (simple)
	bounds := screenshot.GetDisplayBounds(0)
	return screenshot.CaptureRect(bounds)
}
