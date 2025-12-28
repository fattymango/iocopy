//go:build linux
// +build linux

package control

import "runtime"

// BlackScreenWindow manages a fullscreen black window (Linux stub)
type BlackScreenWindow struct{}

// NewBlackScreenWindow creates a new black screen window manager (Linux stub)
func NewBlackScreenWindow() (*BlackScreenWindow, error) {
	if runtime.GOOS != "linux" {
		return nil, nil
	}
	return &BlackScreenWindow{}, nil
}

// Show creates and displays a fullscreen black window (Linux stub - not implemented)
func (b *BlackScreenWindow) Show() error {
	// TODO: Implement X11 fullscreen window for Linux
	return nil
}

// Hide destroys the black screen window (Linux stub)
func (b *BlackScreenWindow) Hide() error {
	return nil
}

// Close cleans up the black screen (Linux stub)
func (b *BlackScreenWindow) Close() error {
	return nil
}

func (b *BlackScreenWindow) GetHotkeyChannel() <-chan struct{} {
	return nil
}

func (b *BlackScreenWindow) OnHotkey() {
	return
}
