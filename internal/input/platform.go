//go:build !linux && !windows
// +build !linux,!windows

package input

import (
	"fmt"
	"runtime"
)

func NewInputCapture() (InputCapture, error) {
	return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func NewInputExecutor() (InputExecutor, error) {
	return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

