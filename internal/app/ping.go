package app

import (
	"copy/internal/wire"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func runPingPong(targetIP, port, localIP string) error {
	client, err := wire.NewClient(targetIP, port)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	log.Printf("[ping-pong] Connected to %s:%s", targetIP, port)

	// Handle incoming messages
	done := make(chan error, 1)
	go func() {
		for {
			msg, err := client.Read()
			if err != nil {
				done <- err
				return
			}
			log.Printf("[ping-pong] Received from %s - Type: %s, Data: %s", targetIP, msg.Type, msg.Data)
		}
	}()

	// Send ping messages
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	msgCount := 0
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			msgCount++
			ping := &wire.Message{
				Type: "ping",
				Data: fmt.Sprintf("ping #%d from %s", msgCount, localIP),
			}
			if err := client.Write(ping); err != nil {
				return fmt.Errorf("write error: %w", err)
			}
			log.Printf("[ping-pong] Sent ping #%d to %s", msgCount, targetIP)
		case err := <-done:
			return err
		case <-sigChan:
			log.Printf("[ping-pong] Interrupted by user")
			return fmt.Errorf("interrupted")
		}
	}
}
