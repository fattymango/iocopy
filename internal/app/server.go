package app

import (
	"copy/internal/control"
	"copy/internal/screencapture"
	"copy/internal/wire"
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
			// Start screen capture and sending raw RGBA frames (matches exp sender protocol on the wire)
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

	w, h := capture.Bounds()
	meta := fmt.Sprintf(`{"width":%d,"height":%d}`, w, h)
	if err := wire.Send(conn, &wire.Message{Type: "screen_stream_meta", Data: meta}); err != nil {
		log.Printf("[screen] Failed to send stream meta: %v", err)
		return
	}
	log.Printf("[screen] Stream dimensions %dx%d for %s", w, h, remoteIP)
	lastW, lastH := w, h

	frameInterval := time.Second / 30
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			log.Printf("[screen] Screen capture stopped for %s", remoteIP)
			return
		case <-ticker.C:
			fw, fh := capture.Bounds()
			if fw != lastW || fh != lastH {
				meta := fmt.Sprintf(`{"width":%d,"height":%d}`, fw, fh)
				if err := wire.Send(conn, &wire.Message{Type: "screen_stream_meta", Data: meta}); err != nil {
					log.Printf("[screen] Failed to send stream meta: %v", err)
					return
				}
				lastW, lastH = fw, fh
				log.Printf("[screen] Stream dimensions updated to %dx%d for %s", fw, fh, remoteIP)
			}
			frameData, err := capture.Capture()
			if err != nil {
				log.Printf("[screen] Failed to capture screen: %v", err)
				continue
			}
			if exp := fw * fh * 4; exp != len(frameData) {
				log.Printf("[screen] Unexpected frame size: got %d want %d", len(frameData), exp)
				continue
			}

			frameMsg := &wire.Message{
				Type: "screen_frame_binary",
				Data: "",
			}
			if err := wire.Send(conn, frameMsg); err != nil {
				log.Printf("[screen] Failed to send frame header: %v", err)
				return
			}
			if err := wire.SendBinaryFrame(conn, frameData); err != nil {
				log.Printf("[screen] Failed to send screen frame: %v", err)
				return
			}
		}
	}
}
