package capture

import "copy/internal/model"

// InputCapture captures local keyboard and mouse input
type InputCapture interface {
	Capture(eventCh chan<- model.InputEvent, stopCh <-chan struct{}) error
	Close() error
}
