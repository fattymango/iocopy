package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"copy/stream"
)

func main() {
	tcpAddr := flag.String("tcp", ":9000", "TCP server address for screen streaming")
	wsAddr := flag.String("ws", ":8080", "WebSocket server address for client connections")
	fps := flag.Int("fps", 25, "Frames per second")
	flag.Parse()

	log.Println("=== Screen Streaming Server ===")
	log.Printf("TCP Server: %s", *tcpAddr)
	log.Printf("WebSocket Server: %s/ws", *wsAddr)
	log.Printf("FPS: %d", *fps)

	// Create streaming pipeline
	s := stream.NewStream(
		*tcpAddr,           // TCP server address
		"localhost"+*tcpAddr, // TCP client connects to localhost
		*wsAddr,            // WebSocket server address
		*fps,               // FPS
	)

	// Start streaming
	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	log.Println("Server started. Waiting for clients...")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	// Stop streaming
	log.Println("Shutting down...")
	s.Stop()
	log.Println("Server stopped")
}




