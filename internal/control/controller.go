package control

import (
	capture "copy/internal/catpure"
	"copy/internal/model"
	"copy/internal/shared"
	"copy/internal/wire"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
)

// Controller captures local input and sends it to the remote peer
type Controller struct {
	client      *wire.Client
	stopCh      chan struct{}
	blackScreen *BlackScreenWindow
}

// NewController creates a new input controller
func NewController(client *wire.Client) *Controller {
	return &Controller{
		client: client,
		stopCh: make(chan struct{}),
	}
}

// Start begins capturing and forwarding input events
func (c *Controller) Start() error {
	log.Printf("[input] Starting input controller...")

	// Create and show black screen window (Windows only)
	if runtime.GOOS == "windows" {
		blackScreen, err := NewBlackScreenWindow()
		if err != nil {
			log.Printf("[input] Warning: Failed to create black screen: %v", err)
		} else {
			c.blackScreen = blackScreen
			if err := blackScreen.Show(); err != nil {
				log.Printf("[input] Warning: Failed to show black screen: %v", err)
			}
		}
		// Ensure black screen is destroyed when done
		defer func() {
			if c.blackScreen != nil {
				c.blackScreen.Close()
			}
		}()
	}

	var capt capture.InputCapture
	var err error

	switch runtime.GOOS {
	case "linux":
		capt, err = capture.NewLinuxInputCapture()
	case "windows":
		capt, err = capture.NewWindowsInputCapture()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	if err != nil {
		return fmt.Errorf("failed to create input capture: %w", err)
	}
	// Start platform-specific input capture
	defer capt.Close()

	log.Printf("[input] Input capture initialized successfully")

	// Channel for input events
	eventCh := make(chan model.InputEvent, 100)

	// Start capturing input in background
	go func() {
		if err := capt.Capture(eventCh, c.stopCh); err != nil {
			log.Printf("[input] Capture error: %v", err)
			// Don't close the channel immediately - keep connection alive
			// The capture might recover or we can continue without it
		}
		// Only close channel if we're actually stopping
		select {
		case <-c.stopCh:
			close(eventCh)
		default:
			// Keep channel open, capture might restart
		}
	}()

	// Send initial control message to establish session
	initMsg := &wire.Message{
		Type: "control_start",
		Data: "Control session started",
	}
	if err := c.client.Write(initMsg); err != nil {
		return fmt.Errorf("failed to send control start message: %w", err)
	}
	log.Printf("[input] Control session established")

	// Start receiving screen frames in background
	frameCh := make(chan string, 60) // Buffer for 60 FPS (1 second of frames)
	go c.receiveScreenFrames(frameCh)

	// Check for hotkey from black screen window (Windows only)
	var hotkeyCh <-chan struct{}
	if runtime.GOOS == "windows" && c.blackScreen != nil {
		hotkeyCh = c.blackScreen.GetHotkeyChannel()
	}

	// Forward events to remote peer and handle screen frames
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed, but don't exit immediately - might be temporary
				log.Printf("[input] Event channel closed, but keeping connection alive")
				// Wait a bit and see if we should stop
				select {
				case <-c.stopCh:
					return nil
				default:
					// Keep waiting for stop signal
					continue
				}
			}

			// Serialize the full event to JSON
			eventData, err := json.Marshal(event)
			if err != nil {
				log.Printf("[input] Failed to marshal event: %v", err)
				continue
			}

			// Check for stop hotkey (Ctrl+Shift+B) - only if not from black screen
			if event.Type == "keyboard" && hotkeyCh == nil {
				var kbEvent model.KeyboardEvent
				if err := json.Unmarshal([]byte(event.Data), &kbEvent); err == nil {
					if kbEvent.Key == "b" &&
						shared.Contains(kbEvent.Modifiers, "ctrl") &&
						shared.Contains(kbEvent.Modifiers, "shift") &&
						kbEvent.Action == "press" {
						log.Printf("[input] Stop hotkey detected (Ctrl+Shift+B)")
						return fmt.Errorf("control stopped by user")
					}
				}
			}

			// Send event to remote peer
			msg := &wire.Message{
				Type: "input_event",
				Data: string(eventData),
			}
			if err := c.client.Write(msg); err != nil {
				log.Printf("[input] Failed to send input event: %v", err)
				return fmt.Errorf("failed to send input event: %w", err)
			}
			log.Printf("[input] Sent input event: %s", event.Type)

		case frameData := <-frameCh:
			// Handle received screen frame - display in black screen window
			if c.blackScreen != nil {
				c.blackScreen.SetFrame(frameData)
			}

		case <-hotkeyCh:
			// Hotkey detected from black screen window
			log.Printf("[input] Stop hotkey detected (Ctrl+Shift+B) from black screen")
			return fmt.Errorf("control stopped by user")

		case <-c.stopCh:
			log.Printf("[input] Controller stopped")
			return nil
		}
	}
}

// receiveScreenFrames receives screen frames from the controlled device
func (c *Controller) receiveScreenFrames(frameCh chan<- string) {
	log.Printf("[screen] Starting to receive screen frames")
	defer close(frameCh)

	for {
		select {
		case <-c.stopCh:
			log.Printf("[screen] Stopping screen frame receiver")
			return
		default:
			// Read message from client (blocking)
			msg, err := c.client.Read()
			if err != nil {
				log.Printf("[screen] Failed to read message: %v", err)
				return
			}

			// Handle screen frame messages
			if msg.Type == "screen_frame" {
				// Frame data is already base64 encoded, send directly
				select {
				case frameCh <- msg.Data:
					// Frame sent successfully
				default:
					// Channel full, skip this frame (non-blocking)
					// This prevents blocking the receiver if display is slow
				}
			} else if msg.Type == "control_ack" {
				log.Printf("[screen] Control acknowledged by remote peer")
			} else {
				log.Printf("[screen] Received unexpected message type: %s", msg.Type)
			}
		}
	}
}

// Stop stops the input controller
func (c *Controller) Stop() {
	close(c.stopCh)
}
