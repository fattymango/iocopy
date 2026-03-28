package ui

import (
	"context"
	"log"
)

// StreamUI handles the streaming client UI
type StreamUI struct {
	ctx     context.Context
	wsAddr  string
	started bool
}

// NewStreamUI creates a new stream UI
func NewStreamUI(wsAddr string) *StreamUI {
	return &StreamUI{
		wsAddr: wsAddr,
	}
}

// Startup is called when Wails starts
func (s *StreamUI) Startup(ctx context.Context) {
	s.ctx = ctx
	log.Println("[stream-ui] UI started")
}

// GetWebSocketAddress returns the WebSocket server address
func (s *StreamUI) GetWebSocketAddress() string {
	return s.wsAddr
}

// StartStream tells the frontend to start connecting to the stream
func (s *StreamUI) StartStream() {
	s.started = true
	log.Println("[stream-ui] Stream start requested")
}

// StopStream tells the frontend to stop the stream
func (s *StreamUI) StopStream() {
	s.started = false
	log.Println("[stream-ui] Stream stop requested")
}

// IsStreamStarted returns whether the stream is started
func (s *StreamUI) IsStreamStarted() bool {
	return s.started
}




