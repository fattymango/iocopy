package app

import (
	"copy/internal/control"
	"copy/internal/wire"
	"fmt"
	"log"
	"net"
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

	// Handle connection immediately (already in a goroutine from server.Start)
	defer func() {
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
