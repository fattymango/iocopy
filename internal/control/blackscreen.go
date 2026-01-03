package control

import "copy/internal/model"

// BlackScreen is an interface for displaying screen frames and handling input events
type BlackScreen interface {
	SetFrame(frameData []byte)
	GetHotkeyChannel() <-chan struct{}
	GetWheelChannel() <-chan model.MouseScrollEvent
	Show() error
	Hide()
}

