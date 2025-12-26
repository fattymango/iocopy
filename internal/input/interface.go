package input

// InputCapture captures local keyboard and mouse input
type InputCapture interface {
	Capture(eventCh chan<- InputEvent, stopCh <-chan struct{}) error
	Close() error
}

// InputExecutor executes keyboard and mouse input on the local system
type InputExecutor interface {
	ExecuteKeyboard(event KeyboardEvent) error
	ExecuteMouseMove(event MouseMoveEvent) error
	ExecuteMouseClick(event MouseClickEvent) error
	ExecuteMouseScroll(event MouseScrollEvent) error
	Close() error
}
