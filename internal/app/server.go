package app

import (
	"copy/internal/control"
	"copy/internal/screencapture"
	"copy/internal/wire"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"time"
)

func (a *App) startServer() error {
	// Create server lazily when actually starting
	if a.server == nil {
		server, err := wire.NewServer(fmt.Sprintf(":%s", a.port))
		if err != nil {
			return fmt.Errorf("failed to start server on port %s: %w. Make sure no other instance is running or use a different port with -port flag", a.port, err)
		}
		a.server = server
	}

	err := a.server.Start(handleServerConnection)
	if err != nil {
		return err
	}
	log.Printf("[app] Server listening on :%s", a.port)
	return nil

}

func handleServerConnection(s *wire.Server, conn net.Conn) {
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	remoteIP := remoteAddr.IP.String()

	log.Printf("[server] Connection accepted from %s", remoteIP)
	log.Printf("[server] Remote peer is taking control of this device")

	// Create input receiver to execute received input events
	receiver, err := control.NewReceiver()
	if err != nil {
		log.Printf("[server] Failed to create input receiver: %v", err)
		conn.Close()
		return
	}
	defer receiver.Close()

	// Screen capture for sending frames back
	screenCapture := screencapture.NewScreenCapture()
	screenCaptureStopCh := make(chan struct{})

	// Handle connection immediately (already in a goroutine from server.Start)
	defer func() {
		close(screenCaptureStopCh)
		conn.Close()
		log.Printf("[server] Connection closed with %s - control session ended", remoteIP)
	}()

	for {
		var msg wire.Message
		if err := wire.Receive(conn, &msg); err != nil {
			log.Printf("[server] Read error from %s: %v", remoteIP, err)
			return
		}

		// Handle different message types
		switch msg.Type {
		case "control_start":
			log.Printf("[server] Control session started by %s", remoteIP)
			// Acknowledge control start
			ack := &wire.Message{
				Type: "control_ack",
				Data: "Control session acknowledged",
			}
			if err := wire.Send(conn, ack); err != nil {
				log.Printf("[server] Failed to send control ack: %v", err)
			}
			// Start screen capture and sending frames
			go startScreenCapture(conn, screenCapture, screenCaptureStopCh, remoteIP)
		case "input_event":
			if err := receiver.HandleMessage(&msg); err != nil {
				log.Printf("[server] Failed to handle input event: %v", err)
				// Continue processing other events
			}
		default:
			log.Printf("[server] Received unknown message type from %s: %s", remoteIP, msg.Type)
		}
	}
}

// startScreenCapture captures the screen and sends frames to the controlling peer
func startScreenCapture(conn net.Conn, capture screencapture.ScreenCapture, stopCh chan struct{}, remoteIP string) {
	log.Printf("[screen] Starting screen capture for %s", remoteIP)
	ticker := time.NewTicker(50 * time.Millisecond) // ~20 FPS
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			log.Printf("[screen] Screen capture stopped for %s", remoteIP)
			return
		case <-ticker.C:
			// Capture screen
			frameData, err := capture.Capture()
			if err != nil {
				log.Printf("[screen] Failed to capture screen: %v", err)
				continue
			}

			// Encode frame as base64 for JSON transmission
			encodedFrame := base64.StdEncoding.EncodeToString(frameData)

			// Send frame to controlling peer
			frameMsg := &wire.Message{
				Type: "screen_frame",
				Data: encodedFrame,
			}
			if err := wire.Send(conn, frameMsg); err != nil {
				log.Printf("[screen] Failed to send screen frame: %v", err)
				return
			}
		}
	}
}
