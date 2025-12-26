package app

import (
	"copy/internal/wire"
	"log"
	"net"
)

func (a *App) startServer() error {
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

	log.Printf("[server] Connection accepted from %s, starting handler", remoteIP)

	// Handle connection immediately (already in a goroutine from server.Start)
	defer func() {
		conn.Close()
		log.Printf("[server] Connection closed with %s", remoteIP)
	}()

	for {
		var msg wire.Message
		if err := wire.Receive(conn, &msg); err != nil {
			log.Printf("[server] Read error from %s: %v", remoteIP, err)
			return
		}
		log.Printf("[server] Received from %s - Type: %s, Data: %s", remoteIP, msg.Type, msg.Data)

		// Ping-pong: echo back immediately
		response := &wire.Message{
			Type: "pong",
			Data: msg.Data,
		}
		if err := wire.Send(conn, response); err != nil {
			log.Printf("[server] Write error to %s: %v", remoteIP, err)
			return
		}
		log.Printf("[server] Sent pong to %s", remoteIP)
	}
}
