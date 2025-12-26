package input

import (
	"copy/internal/wire"
	"encoding/json"
	"fmt"
	"log"
)

// Controller captures local input and sends it to the remote peer
type Controller struct {
	client *wire.Client
	stopCh chan struct{}
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

	// Start platform-specific input capture
	capture, err := NewInputCapture()
	if err != nil {
		log.Printf("[input] Warning: Input capture not available: %v", err)
		log.Printf("[input] Connection will stay alive but input won't be captured")
		// Keep connection alive even without capture
		<-c.stopCh
		return nil
	}
	defer capture.Close()

	log.Printf("[input] Input capture initialized successfully")

	// Channel for input events
	eventCh := make(chan InputEvent, 100)

	// Start capturing input in background
	go func() {
		if err := capture.Capture(eventCh, c.stopCh); err != nil {
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

	// Forward events to remote peer
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

			// Check for stop hotkey (Ctrl+Shift+B)
			if event.Type == "keyboard" {
				var kbEvent KeyboardEvent
				if err := json.Unmarshal([]byte(event.Data), &kbEvent); err == nil {
					if kbEvent.Key == "b" &&
						contains(kbEvent.Modifiers, "ctrl") &&
						contains(kbEvent.Modifiers, "shift") &&
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

		case <-c.stopCh:
			log.Printf("[input] Controller stopped")
			return nil
		}
	}
}

// Stop stops the input controller
func (c *Controller) Stop() {
	close(c.stopCh)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
