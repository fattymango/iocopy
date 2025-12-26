//go:build linux
// +build linux

package input

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func NewInputCapture() (InputCapture, error) {
	return NewLinuxInputCapture()
}

func NewInputExecutor() (InputExecutor, error) {
	return NewLinuxInputExecutor()
}

// LinuxInputCapture captures input using xinput test
type LinuxInputCapture struct {
	stopCh chan struct{}
}

func NewLinuxInputCapture() (InputCapture, error) {
	// Check if xinput is available
	if _, err := exec.LookPath("xinput"); err != nil {
		return nil, fmt.Errorf("xinput not found - please install: sudo apt-get install xinput")
	}
	return &LinuxInputCapture{
		stopCh: make(chan struct{}),
	}, nil
}

func (l *LinuxInputCapture) Capture(eventCh chan<- InputEvent, stopCh <-chan struct{}) error {
	log.Printf("[input] Starting Linux input capture using xinput...")

	// Get keyboard and mouse device IDs
	keyboardID, err := getKeyboardID()
	if err != nil {
		log.Printf("[input] Warning: Could not find keyboard device, keyboard capture may not work: %v", err)
	}

	mouseID, err := getMouseID()
	if err != nil {
		log.Printf("[input] Warning: Could not find mouse device, mouse capture may not work: %v", err)
	}

	// Start keyboard capture
	if keyboardID != "" {
		go l.captureKeyboard(keyboardID, eventCh, stopCh)
	}

	// Start mouse capture
	if mouseID != "" {
		go l.captureMouse(mouseID, eventCh, stopCh)
	}

	// Keep running until stopped
	<-stopCh
	log.Printf("[input] Input capture stopped")
	return nil
}

func (l *LinuxInputCapture) captureKeyboard(deviceID string, eventCh chan<- InputEvent, stopCh <-chan struct{}) {
	cmd := exec.Command("xinput", "test", deviceID)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[input] Failed to create stdout pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[input] Failed to start xinput: %v", err)
		return
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-stopCh:
			cmd.Process.Kill()
			return
		default:
		}

		line := scanner.Text()
		if strings.Contains(line, "key press") || strings.Contains(line, "key release") {
			// Parse xinput output: "key press   54" or "key release 54"
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				action := "press"
				if parts[1] == "release" {
					action = "release"
				}
				keyCode := parts[2]

				// Convert key code to key name (simplified)
				keyName := keyCodeToName(keyCode)

				kbEvent := KeyboardEvent{
					Key:       keyName,
					Action:    action,
					Modifiers: []string{}, // xinput test doesn't show modifiers easily
				}

				data, _ := json.Marshal(kbEvent)
				eventCh <- InputEvent{
					Type: "keyboard",
					Data: string(data),
				}
			}
		}
	}
}

func (l *LinuxInputCapture) captureMouse(deviceID string, eventCh chan<- InputEvent, stopCh <-chan struct{}) {
	cmd := exec.Command("xinput", "test-xi2", "--root")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[input] Failed to create mouse stdout pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[input] Failed to start xinput for mouse: %v", err)
		return
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	var lastX, lastY int
	for scanner.Scan() {
		select {
		case <-stopCh:
			cmd.Process.Kill()
			return
		default:
		}

		line := scanner.Text()
		// Parse mouse events from xinput test-xi2
		// This is simplified - real implementation would parse XI2 events properly
		if strings.Contains(line, "EVENT") {
			// For now, get current mouse position
			x, y, err := getMousePosition()
			if err == nil && (x != lastX || y != lastY) {
				moveEvent := MouseMoveEvent{X: x, Y: y}
				data, _ := json.Marshal(moveEvent)
				eventCh <- InputEvent{
					Type: "mouse_move",
					Data: string(data),
				}
				lastX, lastY = x, y
			}
		}
	}
}

func getKeyboardID() (string, error) {
	// Try to find a keyboard device
	cmd2 := exec.Command("xinput", "list")
	output2, err := cmd2.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output2), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "keyboard") && !strings.Contains(strings.ToLower(line), "virtual") {
			// Extract ID (first number in line)
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
		}
	}

	// Fallback: use first device
	if len(lines) > 0 {
		fields := strings.Fields(lines[0])
		if len(fields) > 0 {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("no keyboard device found")
}

func getMouseID() (string, error) {
	cmd := exec.Command("xinput", "list")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if (strings.Contains(strings.ToLower(line), "mouse") || strings.Contains(strings.ToLower(line), "pointer")) &&
			!strings.Contains(strings.ToLower(line), "virtual") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
		}
	}

	return "", fmt.Errorf("no mouse device found")
}

func getMousePosition() (int, int, error) {
	cmd := exec.Command("xdotool", "getmouselocation", "--shell")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var x, y int
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "X=") {
			fmt.Sscanf(line, "X=%d", &x)
		}
		if strings.HasPrefix(line, "Y=") {
			fmt.Sscanf(line, "Y=%d", &y)
		}
	}

	return x, y, nil
}

func keyCodeToName(code string) string {
	// Simplified key code mapping - in production, use proper X11 keycode mapping
	keyMap := map[string]string{
		"36": "Return", "37": "Control_L", "50": "Shift_L", "64": "Alt_L",
		"9": "Escape", "23": "Tab", "65": "space", "22": "BackSpace",
	}
	if name, ok := keyMap[code]; ok {
		return name
	}
	// For letters, try to map
	if code >= "10" && code <= "35" {
		// Rough mapping for a-z
		return fmt.Sprintf("key_%s", code)
	}
	return code
}

func (l *LinuxInputCapture) Close() error {
	return nil
}

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
