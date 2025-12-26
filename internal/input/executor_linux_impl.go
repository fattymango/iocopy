package input

import (
	"fmt"
	"os/exec"
)

// LinuxInputExecutor executes input using xdotool
type LinuxInputExecutor struct{}

func NewLinuxInputExecutor() (InputExecutor, error) {
	// Check if xdotool is available
	if _, err := exec.LookPath("xdotool"); err != nil {
		return nil, fmt.Errorf("xdotool not found - please install: sudo apt-get install xdotool")
	}
	return &LinuxInputExecutor{}, nil
}

func (l *LinuxInputExecutor) ExecuteKeyboard(event KeyboardEvent) error {
	// Build xdotool command
	var cmd *exec.Cmd

	if event.Action == "press" {
		// Press key
		key := event.Key
		if len(event.Modifiers) > 0 {
			// Combine modifiers
			mods := ""
			for _, mod := range event.Modifiers {
				mods += mod + "+"
			}
			key = mods + key
		}
		cmd = exec.Command("xdotool", "key", key)
	} else {
		// Release key
		cmd = exec.Command("xdotool", "keyup", event.Key)
	}

	return cmd.Run()
}

func (l *LinuxInputExecutor) ExecuteMouseMove(event MouseMoveEvent) error {
	cmd := exec.Command("xdotool", "mousemove", fmt.Sprintf("%d", event.X), fmt.Sprintf("%d", event.Y))
	return cmd.Run()
}

func (l *LinuxInputExecutor) ExecuteMouseClick(event MouseClickEvent) error {
	var cmd *exec.Cmd

	button := "1" // left
	switch event.Button {
	case "right":
		button = "3"
	case "middle":
		button = "2"
	}

	if event.Action == "press" {
		cmd = exec.Command("xdotool", "mousedown", button)
	} else {
		cmd = exec.Command("xdotool", "mouseup", button)
	}

	return cmd.Run()
}

func (l *LinuxInputExecutor) ExecuteMouseScroll(event MouseScrollEvent) error {
	// xdotool doesn't have direct scroll, use click with button 4/5
	if event.DeltaY > 0 {
		cmd := exec.Command("xdotool", "click", "4")
		return cmd.Run()
	} else if event.DeltaY < 0 {
		cmd := exec.Command("xdotool", "click", "5")
		return cmd.Run()
	}
	return nil
}

func (l *LinuxInputExecutor) Close() error {
	return nil
}
