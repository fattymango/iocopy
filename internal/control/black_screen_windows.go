//go:build windows
// +build windows

package control

import (
	"context"
	"embed"
	"io/fs"
	"log"
	gosync "runtime"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:blackscreen
var blackScreenAssets embed.FS

// BlackScreenApp is the Wails app struct for handling hotkey and displaying frames
type BlackScreenApp struct {
	ctx      context.Context
	hotkeyCh chan struct{}
}

// OnHotkey is called from JavaScript when Ctrl+Shift+B is pressed
func (a *BlackScreenApp) OnHotkey() {
	log.Printf("[blackscreen] Hotkey pressed")
	select {
	case a.hotkeyCh <- struct{}{}:
	default:
	}
}

// SetFrame is called from Go to update the displayed frame
func (a *BlackScreenApp) SetFrame(frameData string) {
	if a.ctx == nil {
		return
	}
	// Emit event to JavaScript with base64 frame data
	wailsruntime.EventsEmit(a.ctx, "frame", frameData)
}

// BlackScreenWindow manages a fullscreen black window using Wails
type BlackScreenWindow struct {
	app      *BlackScreenApp
	wg       sync.WaitGroup
	stopCh   chan struct{}
	hotkeyCh chan struct{} // Channel to signal hotkey detected
	running  bool
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// SetFrame sets the frame data to display in the black screen window
func (b *BlackScreenWindow) SetFrame(frameData string) {
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
	}, nil
}

// GetHotkeyChannel returns the channel that signals when stop hotkey is pressed
func (b *BlackScreenWindow) GetHotkeyChannel() <-chan struct{} {
	return b.hotkeyCh
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
		hotkeyCh: b.hotkeyCh,
	}

	b.mu.Lock()
	b.app = app
	b.mu.Unlock()

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
	return nil
}
