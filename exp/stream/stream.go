package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
)

type Stream struct {
	mu    sync.RWMutex
	frame []byte
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) Start(addr string) {
	go func() {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal("stream listen error:", err)
		}
		log.Println("[stream] listening on", addr)

		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("stream accept error:", err)
		}
		defer conn.Close()

		log.Println("[stream] sender connected")

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

			s.mu.Lock()
			s.frame = buf
			s.mu.Unlock()

			log.Printf("[stream] frame received: %d bytes\n", size)
		}
	}()
}

// Wails binding
func (s *Stream) GetLatestFrame() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.frame
}
