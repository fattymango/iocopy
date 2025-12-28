package control

import (
	"copy/internal/executor"
	"copy/internal/model"
	"copy/internal/wire"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
)

// Receiver receives input events and executes them locally
type Receiver struct {
	executor executor.InputExecutor
}

// NewReceiver creates a new input receiver
func NewReceiver() (*Receiver, error) {
	var execu executor.InputExecutor
	var err error

	switch runtime.GOOS {
	case "linux":
		execu, err = executor.NewLinuxInputExecutor()
	case "windows":
		execu, err = executor.NewWindowsInputExecutor()
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create input executor: %w", err)
	}

	return &Receiver{
		executor: execu,
	}, nil
}

// HandleMessage processes an input event message and executes it
func (r *Receiver) HandleMessage(msg *wire.Message) error {
	if msg.Type != "input_event" {
		return nil // Not an input event, ignore
	}

	var event model.InputEvent
	if err := json.Unmarshal([]byte(msg.Data), &event); err != nil {
		return fmt.Errorf("failed to unmarshal input event: %w", err)
	}

	log.Printf("[input] Received event: Type=%s", event.Type)

	switch event.Type {
	case "keyboard":
		var kbEvent model.KeyboardEvent
		if err := json.Unmarshal([]byte(event.Data), &kbEvent); err != nil {
			return fmt.Errorf("failed to unmarshal keyboard event: %w", err)
		}
		return r.executor.ExecuteKeyboard(kbEvent)

	case "mouse_move":
		var moveEvent model.MouseMoveEvent
		if err := json.Unmarshal([]byte(event.Data), &moveEvent); err != nil {
			return fmt.Errorf("failed to unmarshal mouse move event: %w", err)
		}
		return r.executor.ExecuteMouseMove(moveEvent)

	case "mouse_click":
		var clickEvent model.MouseClickEvent
		if err := json.Unmarshal([]byte(event.Data), &clickEvent); err != nil {
			return fmt.Errorf("failed to unmarshal mouse click event: %w", err)
		}
		return r.executor.ExecuteMouseClick(clickEvent)

	case "mouse_scroll":
		var scrollEvent model.MouseScrollEvent
		if err := json.Unmarshal([]byte(event.Data), &scrollEvent); err != nil {
			return fmt.Errorf("failed to unmarshal mouse scroll event: %w", err)
		}
		return r.executor.ExecuteMouseScroll(scrollEvent)

	default:
		log.Printf("[input] Unknown event type: %s", event.Type)
		return nil
	}
}

// Close closes the receiver
func (r *Receiver) Close() error {
	if r.executor != nil {
		return r.executor.Close()
	}
	return nil
}
