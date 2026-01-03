package ui

import (
	"context"
	"copy/internal/model"
	"copy/internal/shared"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type UI struct {
	app  Application
	ctx  context.Context
	port string

	// Blackscreen functionality
	mu                sync.Mutex
	hotkeyCh          chan struct{}
	wheelCh           chan model.MouseScrollEvent
	wsConn            *websocket.Conn
	wsMu              sync.Mutex
	frameQueue        chan []byte
	stopFrame         chan struct{}
	wsServer          *http.Server
	wsUpgrader        websocket.Upgrader
	blackScreenActive bool
}

func NewUI(app Application, port string) *UI {
	return &UI{
		app:        app,
		port:       port,
		hotkeyCh:   make(chan struct{}, 1),
		wheelCh:    make(chan model.MouseScrollEvent, 10),
		frameQueue: make(chan []byte, 2),
		stopFrame:  make(chan struct{}),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from localhost only
			},
		},
	}
}

func (u *UI) Startup(ctx context.Context) {
	u.ctx = ctx
	log.Println("[ui] UI started")
	log.Printf("[ui] UI methods available: OnHotkey=%v, OnMouseWheel=%v, ShowBlackScreen=%v, HideBlackScreen=%v",
		u != nil, u != nil, u != nil, u != nil)
}

// Called by frontend to scan peers
func (u *UI) ScanPeers() []string {
	log.Println("[ui] Scanning for peers...")
	return u.app.FindReachableIPs(u.port)
}

// Called by frontend when user selects an IP
func (u *UI) Connect(ip string) error {
	log.Printf("[ui] Connecting to %s:%s", ip, u.port)
	return u.app.RunControl(ip, u.port)
}

// Optional: expose local IP
func (u *UI) LocalIP() string {
	return shared.GetLocalIP()
}

// OnHotkey is called from JavaScript when Ctrl+Shift+B is pressed
func (u *UI) OnHotkey() {
	log.Printf("[ui] Hotkey pressed")
	select {
	case u.hotkeyCh <- struct{}{}:
		log.Printf("[ui] Hotkey sent to channel")
	default:
		log.Printf("[ui] WARNING: Hotkey channel full, dropping event")
	}
}

// OnMouseWheel is called from JavaScript when mouse wheel is scrolled
func (u *UI) OnMouseWheel(deltaX, deltaY int) {
	scrollEvent := model.MouseScrollEvent{
		DeltaX: deltaX,
		DeltaY: deltaY,
	}
	log.Printf("[ui] Mouse wheel: deltaX=%d, deltaY=%d", deltaX, deltaY)
	select {
	case u.wheelCh <- scrollEvent:
		log.Printf("[ui] Wheel event sent to channel")
	default:
		log.Printf("[ui] WARNING: Wheel channel full, dropping event")
	}
}

// SetFrame sets the frame data to display in the black screen
func (u *UI) SetFrame(frameData []byte) {
	select {
	case u.frameQueue <- frameData:
		// Frame queued successfully
	default:
		// Queue full, drop frame to prevent blocking
	}
}

// GetHotkeyChannel returns the channel that signals when stop hotkey is pressed
func (u *UI) GetHotkeyChannel() <-chan struct{} {
	log.Printf("[ui] GetHotkeyChannel called, channel exists: %v", u.hotkeyCh != nil)
	return u.hotkeyCh
}

// GetWheelChannel returns the channel that receives mouse wheel events
func (u *UI) GetWheelChannel() <-chan model.MouseScrollEvent {
	log.Printf("[ui] GetWheelChannel called, channel exists: %v", u.wheelCh != nil)
	return u.wheelCh
}

// ShowBlackScreen starts the blackscreen view and WebSocket server (called from frontend)
func (u *UI) ShowBlackScreen() error {
	return u.Show()
}

// HideBlackScreen stops the blackscreen view (called from frontend)
func (u *UI) HideBlackScreen() {
	u.Hide()
}

// Show starts the blackscreen view and WebSocket server (implements control.BlackScreen)
func (u *UI) Show() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.blackScreenActive {
		return nil
	}

	u.blackScreenActive = true

	// Start frame sender goroutine
	go u.startFrameSender()

	// Start WebSocket server
	u.startWebSocketServer()

	// Make window fullscreen
	if u.ctx != nil {
		wailsruntime.WindowFullscreen(u.ctx)
		log.Printf("[ui] Window set to fullscreen")
	}

	return nil
}

// Hide stops the blackscreen view (implements control.BlackScreen)
func (u *UI) Hide() {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.blackScreenActive {
		return
	}

	u.blackScreenActive = false

	// Exit fullscreen
	if u.ctx != nil {
		wailsruntime.WindowUnfullscreen(u.ctx)
		log.Printf("[ui] Window exited fullscreen")
	}

	// Stop frame sender
	select {
	case <-u.stopFrame:
		// Already closed
	default:
		close(u.stopFrame)
	}

	// Close WebSocket connection if open
	u.wsMu.Lock()
	if u.wsConn != nil {
		u.wsConn.Close()
		u.wsConn = nil
	}
	u.wsMu.Unlock()

	// Stop WebSocket server
	if u.wsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		u.wsServer.Shutdown(ctx)
		u.wsServer = nil
	}
}

// startFrameSender starts a goroutine that sends frames over WebSocket
func (u *UI) startFrameSender() {
	// Recreate stopFrame channel
	u.mu.Lock()
	u.stopFrame = make(chan struct{})
	u.mu.Unlock()

	go func() {
		for {
			select {
			case <-u.stopFrame:
				return
			case frameData := <-u.frameQueue:
				u.wsMu.Lock()
				conn := u.wsConn
				u.wsMu.Unlock()

				if conn != nil {
					// Set write deadline to prevent blocking (100ms timeout)
					conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))

					// Send frame data as binary WebSocket message
					if err := conn.WriteMessage(websocket.BinaryMessage, frameData); err != nil {
						log.Printf("[ui] Failed to send frame over WebSocket: %v", err)
						// Clear connection on error
						u.wsMu.Lock()
						u.wsConn = nil
						u.wsMu.Unlock()
					}
				}
			}
		}
	}()
}

// startWebSocketServer starts a WebSocket server on port 8765 for frame streaming
func (u *UI) startWebSocketServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/frames", func(w http.ResponseWriter, r *http.Request) {
		conn, err := u.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[ui] WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		log.Printf("[ui] WebSocket client connected")

		// Store connection
		u.wsMu.Lock()
		u.wsConn = conn
		u.wsMu.Unlock()

		// Keep connection alive and handle close
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[ui] WebSocket connection closed: %v", err)
				// Clear connection
				u.wsMu.Lock()
				u.wsConn = nil
				u.wsMu.Unlock()
				break
			}
		}
	})

	u.wsServer = &http.Server{
		Addr:    "127.0.0.1:8765",
		Handler: mux,
	}

	go func() {
		log.Printf("[ui] Starting WebSocket server on port 8765")
		if err := u.wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ui] WebSocket server error: %v", err)
		}
	}()
}
