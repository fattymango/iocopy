//go:build windows
// +build windows

package control

import (
	"context"
	"copy/internal/model"
	"embed"
	"io/fs"
	"log"
	"net/http"
	gosync "runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	// WebSocket server port for frame streaming
	websocketPort = "8765"
)

//go:embed all:blackscreen
var blackScreenAssets embed.FS

// BlackScreenApp is the Wails app struct for handling hotkey and displaying frames
type BlackScreenApp struct {
	ctx        context.Context
	hotkeyCh   chan struct{}
	wheelCh    chan model.MouseScrollEvent
	wsConn     *websocket.Conn
	wsMu       sync.Mutex
	frameQueue chan []byte // Use bytes instead of base64 string
	stopFrame  chan struct{}
}

// OnHotkey is called from JavaScript when Ctrl+Shift+B is pressed
func (a *BlackScreenApp) OnHotkey() {
	log.Printf("[blackscreen] Hotkey pressed")
	select {
	case a.hotkeyCh <- struct{}{}:
	default:
	}
}

// OnMouseWheel is called from JavaScript when mouse wheel is scrolled
func (a *BlackScreenApp) OnMouseWheel(deltaX, deltaY int) {
	scrollEvent := model.MouseScrollEvent{
		DeltaX: deltaX,
		DeltaY: deltaY,
	}
	select {
	case a.wheelCh <- scrollEvent:
	default:
		// Channel full, drop event
	}
}

// SetFrame is called from Go to update the displayed frame (non-blocking)
func (a *BlackScreenApp) SetFrame(frameData []byte) {
	select {
	case a.frameQueue <- frameData:
		// Frame queued successfully
	default:
		// Queue full, drop frame to prevent blocking
		// This prevents lag when browser can't keep up
	}
}

// startFrameSender starts a goroutine that sends frames over WebSocket
func (a *BlackScreenApp) startFrameSender() {
	go func() {
		for {
			select {
			case <-a.stopFrame:
				return
			case frameData := <-a.frameQueue:
				a.wsMu.Lock()
				conn := a.wsConn
				a.wsMu.Unlock()

				if conn != nil {
					// Set write deadline to prevent blocking (100ms timeout)
					conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))

					// Send frame data as binary WebSocket message (more efficient than base64 text)
					if err := conn.WriteMessage(websocket.BinaryMessage, frameData); err != nil {
						log.Printf("[blackscreen] Failed to send frame over WebSocket: %v", err)
						// Clear connection on error
						a.wsMu.Lock()
						a.wsConn = nil
						a.wsMu.Unlock()
					}
				}
			}
		}
	}()
}

// BlackScreenWindow manages a fullscreen black window using Wails
type BlackScreenWindow struct {
	app        *BlackScreenApp
	wg         sync.WaitGroup
	stopCh     chan struct{}
	hotkeyCh   chan struct{}               // Channel to signal hotkey detected
	wheelCh    chan model.MouseScrollEvent // Channel for wheel events
	running    bool
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wsServer   *http.Server
	wsUpgrader websocket.Upgrader
}

// SetFrame sets the frame data to display in the black screen window
func (b *BlackScreenWindow) SetFrame(frameData []byte) {
	b.mu.Lock()
	app := b.app
	b.mu.Unlock()

	if app != nil {
		app.SetFrame(frameData)
	}
}

// NewBlackScreenWindow creates a new black screen window manager
func NewBlackScreenWindow() (*BlackScreenWindow, error) {
	return &BlackScreenWindow{
		stopCh:   make(chan struct{}),
		hotkeyCh: make(chan struct{}, 1),
		wheelCh:  make(chan model.MouseScrollEvent, 10),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from localhost only
				return true
			},
		},
	}, nil
}

// GetHotkeyChannel returns the channel that signals when stop hotkey is pressed
func (b *BlackScreenWindow) GetHotkeyChannel() <-chan struct{} {
	return b.hotkeyCh
}

// GetWheelChannel returns the channel that receives mouse wheel events
func (b *BlackScreenWindow) GetWheelChannel() <-chan model.MouseScrollEvent {
	return b.wheelCh
}

// Show creates and displays a fullscreen black window
func (b *BlackScreenWindow) Show() error {
	if gosync.GOOS != "windows" {
		return nil
	}

	log.Printf("[blackscreen] Creating fullscreen black window...")

	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	// Create app context
	ctx, cancel := context.WithCancel(context.Background())
	b.ctx = ctx
	b.cancel = cancel

	// Create app struct
	app := &BlackScreenApp{
		hotkeyCh:   b.hotkeyCh,
		wheelCh:    b.wheelCh,
		frameQueue: make(chan []byte, 2), // Small buffer - drop frames if browser can't keep up
		stopFrame:  make(chan struct{}),
	}

	// Start frame sender goroutine
	app.startFrameSender()

	b.mu.Lock()
	b.app = app
	b.mu.Unlock()

	// Start WebSocket server for frame streaming
	b.startWebSocketServer()

	// Run Wails app in a goroutine
	b.wg.Add(1)
	go func() {
		defer func() {
			b.mu.Lock()
			b.running = false
			// send hotkey event to the app to stop the app
			if b.app != nil {
				b.app.hotkeyCh <- struct{}{}
			}
			b.app = nil
			b.mu.Unlock()
			b.wg.Done()
			log.Printf("[blackscreen] Wails app exited")
		}()

		b.mu.Lock()
		b.running = true
		b.mu.Unlock()

		log.Printf("[blackscreen] Black screen displayed (fullscreen)")

		// Create asset server with embedded HTML
		htmlFS, err := fs.Sub(blackScreenAssets, "blackscreen")
		if err != nil {
			log.Printf("[blackscreen] Failed to create sub FS: %v", err)
			htmlFS = blackScreenAssets
		}

		assetServer := &assetserver.Options{
			Assets: htmlFS,
		}

		// Run the Wails app (this blocks until window is closed)
		// Pass the app struct in Bind so Wails can bind it for JavaScript access
		// Note: Wails needs to be built with wails build command to work properly
		err = wails.Run(&options.App{
			Title:     "IOCopy Black Screen",
			Width:     1920,
			Height:    1080,
			MinWidth:  800,
			MinHeight: 600,

			// ✅ IMPORTANT
			Frameless:   false, // must be false for native title bar
			AlwaysOnTop: true,

			AssetServer:      assetServer,
			BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 255},
			Bind:             []interface{}{app},

			// Enable GPU acceleration for WebView2 (set to false to enable GPU)
			Windows: &windows.Options{
				WebviewGpuIsDisabled: false, // Enable GPU acceleration
			},

			OnStartup: func(ctx context.Context) {
				app.ctx = ctx

				wailsruntime.WindowShow(ctx)
				wailsruntime.WindowSetAlwaysOnTop(ctx, true)

				// Enter fullscreen mode to take the entire window
				wailsruntime.WindowFullscreen(ctx)

				log.Printf("[blackscreen] Window started in fullscreen mode")
			},
			OnShutdown: func(ctx context.Context) {
				log.Printf("[blackscreen] Window shutdown")
			},
		})

		if err != nil {
			log.Printf("[blackscreen] Wails Run error: %v", err)
			// If Wails fails, log the error but don't crash
		}
	}()

	// Give app a moment to initialize
	time.Sleep(200 * time.Millisecond)

	return nil
}

// Hide destroys the black screen window
func (b *BlackScreenWindow) Hide() error {
	if gosync.GOOS != "windows" {
		return nil
	}

	b.mu.Lock()
	if !b.running || b.app == nil {
		b.mu.Unlock()
		return nil
	}
	app := b.app
	b.running = false

	// Stop frame sender
	select {
	case <-app.stopFrame:
		// Already closed
	default:
		close(app.stopFrame)
	}

	// Close WebSocket connection if open
	if app.wsConn != nil {
		app.wsMu.Lock()
		if app.wsConn != nil {
			app.wsConn.Close()
			app.wsConn = nil
		}
		app.wsMu.Unlock()
	}
	b.mu.Unlock()

	log.Printf("[blackscreen] Destroying black screen window...")

	// Close the window using runtime
	if app.ctx != nil {
		wailsruntime.Quit(app.ctx)
	}

	// Cancel context to stop the app
	if b.cancel != nil {
		b.cancel()
	}

	log.Printf("[blackscreen] Black screen window destroyed")

	return nil
}

// Close cleans up the black screen
func (b *BlackScreenWindow) Close() error {
	b.Hide()
	b.wg.Wait()

	// Stop WebSocket server
	if b.wsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		b.wsServer.Shutdown(ctx)
	}

	return nil
}

// startWebSocketServer starts a WebSocket server on a fixed port for frame streaming
func (b *BlackScreenWindow) startWebSocketServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/frames", func(w http.ResponseWriter, r *http.Request) {
		conn, err := b.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[blackscreen] WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		log.Printf("[blackscreen] WebSocket client connected")

		// Store connection in app
		b.mu.Lock()
		if b.app != nil {
			b.app.wsMu.Lock()
			b.app.wsConn = conn
			b.app.wsMu.Unlock()
		}
		b.mu.Unlock()

		// Keep connection alive and handle close
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[blackscreen] WebSocket connection closed: %v", err)
				// Clear connection
				b.mu.Lock()
				if b.app != nil {
					b.app.wsMu.Lock()
					b.app.wsConn = nil
					b.app.wsMu.Unlock()
				}
				b.mu.Unlock()
				break
			}
		}
	})

	b.wsServer = &http.Server{
		Addr:    "127.0.0.1:" + websocketPort,
		Handler: mux,
	}

	go func() {
		log.Printf("[blackscreen] Starting WebSocket server on port %s", websocketPort)
		if err := b.wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[blackscreen] WebSocket server error: %v", err)
		}
	}()
}
