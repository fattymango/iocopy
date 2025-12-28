package capture

import (
	"copy/internal/model"
	"copy/pkg/windows"
	"encoding/json"
	"log"
	"runtime"
	"time"
	"unsafe"
)

// Windows DLL variables are declared in windows_dll.go (with build tags)
// This allows the code to compile on all platforms without build tags in this file

type POINT struct {
	X, Y int32
}

type MSLLHOOKSTRUCT struct {
	Pt          POINT
	MouseData   uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// WindowsInputCapture captures input on Windows using low-level hooks
type WindowsInputCapture struct {
	stopCh       chan struct{}
	keyboardHook uintptr
	mouseHook    uintptr
}

func NewWindowsInputCapture() (InputCapture, error) {
	if err := windows.InitWindowsDLLs(); err != nil {
		return nil, err
	}
	return &WindowsInputCapture{
		stopCh: make(chan struct{}),
	}, nil
}

func (w *WindowsInputCapture) Capture(eventCh chan<- model.InputEvent, stopCh <-chan struct{}) error {
	log.Printf("[input] Starting Windows input capture...")

	// Start keyboard capture
	go w.captureKeyboard(eventCh, stopCh)

	// Start mouse capture
	go w.captureMouse(eventCh, stopCh)

	// Keep running until stopped
	<-stopCh
	log.Printf("[input] Windows input capture stopped")
	return nil
}

func (w *WindowsInputCapture) captureKeyboard(eventCh chan<- model.InputEvent, stopCh <-chan struct{}) {
	if runtime.GOOS != "windows" {
		return
	}

	// Use GetAsyncKeyState in a loop to capture all keys
	// When blocking, we consume the input but don't let it reach the system
	keyStates := make(map[uint32]bool)

	// Map of virtual key codes to key names (0-255)
	keyNames := make(map[uint32]string)
	for i := uint32(0x41); i <= 0x5A; i++ { // A-Z
		keyNames[i] = string(rune('a' + (i - 0x41)))
	}
	keyNames[windows.VK_CONTROL] = "Control_L"
	keyNames[windows.VK_SHIFT] = "Shift_L"
	keyNames[0x0D] = "Return"
	keyNames[0x1B] = "Escape"
	keyNames[0x09] = "Tab"
	keyNames[0x20] = "space"
	keyNames[0x08] = "BackSpace"
	// Add more as needed

	for {
		select {
		case <-stopCh:
			return
		default:
			// Check all keys in the map
			for vkCode, keyName := range keyNames {
				state, _, _ := windows.ProcGetAsyncKeyState.Call(uintptr(vkCode))
				isPressed := (state & 0x8000) != 0

				wasPressed := keyStates[vkCode]
				if isPressed != wasPressed {
					keyStates[vkCode] = isPressed

					// Check modifiers (always check current state)
					ctrlState, _, _ := windows.ProcGetAsyncKeyState.Call(uintptr(windows.VK_CONTROL))
					shiftState, _, _ := windows.ProcGetAsyncKeyState.Call(uintptr(windows.VK_SHIFT))
					altState, _, _ := windows.ProcGetAsyncKeyState.Call(uintptr(0x12)) // VK_MENU (Alt)

					ctrlPressed := (ctrlState & 0x8000) != 0
					shiftPressed := (shiftState & 0x8000) != 0
					altPressed := (altState & 0x8000) != 0

					modifiers := []string{}
					if ctrlPressed {
						modifiers = append(modifiers, "ctrl")
					}
					if shiftPressed {
						modifiers = append(modifiers, "shift")
					}
					if altPressed {
						modifiers = append(modifiers, "alt")
					}

					// Only send events for actual key presses (not just modifier changes)
					if vkCode != windows.VK_CONTROL && vkCode != windows.VK_SHIFT && vkCode != 0x12 {
						kbEvent := model.KeyboardEvent{
							Key:       keyName,
							Action:    "press",
							Modifiers: modifiers,
						}
						if !isPressed {
							kbEvent.Action = "release"
						}

						data, _ := json.Marshal(kbEvent)
						eventCh <- model.InputEvent{
							Type: "keyboard",
							Data: string(data),
						}
					}
				}
			}

			// Small delay to avoid CPU spinning
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (w *WindowsInputCapture) captureMouse(eventCh chan<- model.InputEvent, stopCh <-chan struct{}) {
	if runtime.GOOS != "windows" {
		return
	}

	var lastX, lastY int32
	buttonStates := make(map[uint32]bool)          // Track button press states
	lastClickTime := make(map[uint32]int64)        // Track last click time for double-click detection
	lastClickPos := make(map[uint32]POINT)         // Track last click position
	inDoubleClickSequence := make(map[uint32]bool) // Track if we're in a double-click sequence

	// Button mappings
	buttons := map[uint32]string{
		windows.VK_LBUTTON: "left",
		windows.VK_RBUTTON: "right",
		windows.VK_MBUTTON: "middle",
	}

	for {
		select {
		case <-stopCh:
			return
		default:
			var pt POINT
			windows.ProcGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

			// Check mouse movement
			if pt.X != lastX || pt.Y != lastY {
				moveEvent := model.MouseMoveEvent{
					X: int(pt.X),
					Y: int(pt.Y),
				}
				data, _ := json.Marshal(moveEvent)
				eventCh <- model.InputEvent{
					Type: "mouse_move",
					Data: string(data),
				}
				lastX, lastY = pt.X, pt.Y
			}

			// Check mouse buttons
			now := time.Now().UnixNano()
			doubleClickDelay := int64(500 * time.Millisecond) // Windows default double-click time (~500ms)
			maxDoubleClickDistance := int32(5)                // Maximum pixels for double-click detection

			for vkCode, buttonName := range buttons {
				state, _, _ := windows.ProcGetAsyncKeyState.Call(uintptr(vkCode))
				isPressed := (state & 0x8000) != 0
				wasPressed := buttonStates[vkCode]

				if isPressed != wasPressed {
					buttonStates[vkCode] = isPressed

					if isPressed {
						// Button pressed
						// Check for double click (second press within time/distance window)
						lastTime, hadPreviousClick := lastClickTime[vkCode]
						lastPos, hadPreviousPos := lastClickPos[vkCode]
						isDoubleClick := false

						if hadPreviousClick && hadPreviousPos {
							timeSinceLastClick := now - lastTime
							distanceX := pt.X - lastPos.X
							if distanceX < 0 {
								distanceX = -distanceX
							}
							distanceY := pt.Y - lastPos.Y
							if distanceY < 0 {
								distanceY = -distanceY
							}

							if timeSinceLastClick < doubleClickDelay &&
								distanceX < maxDoubleClickDistance &&
								distanceY < maxDoubleClickDistance {
								isDoubleClick = true
								inDoubleClickSequence[vkCode] = true // Mark that we're in a double-click sequence
							}
						}

						if isDoubleClick {
							// Send double-click event (executor will send full press/release/press/release sequence)
							clickEvent := model.MouseClickEvent{
								Button:   buttonName,
								Action:   "double",
								X:        int(pt.X),
								Y:        int(pt.Y),
								IsDouble: true,
							}
							data, _ := json.Marshal(clickEvent)
							eventCh <- model.InputEvent{
								Type: "mouse_click",
								Data: string(data),
							}
						} else {
							// Normal single click press
							clickEvent := model.MouseClickEvent{
								Button:   buttonName,
								Action:   "press",
								X:        int(pt.X),
								Y:        int(pt.Y),
								IsDouble: false,
							}
							data, _ := json.Marshal(clickEvent)
							eventCh <- model.InputEvent{
								Type: "mouse_click",
								Data: string(data),
							}
						}

						lastClickTime[vkCode] = now
						lastClickPos[vkCode] = pt
					} else {
						// Button released
						// If we're in a double-click sequence, skip sending the release
						// because the executor already sends press/release/press/release for double clicks
						if inDoubleClickSequence[vkCode] {
							inDoubleClickSequence[vkCode] = false // Reset flag
							// Don't send release event for double-click sequence
							continue
						}

						// Normal single click release
						clickEvent := model.MouseClickEvent{
							Button:   buttonName,
							Action:   "release",
							X:        int(pt.X),
							Y:        int(pt.Y),
							IsDouble: false,
						}

						data, _ := json.Marshal(clickEvent)
						eventCh <- model.InputEvent{
							Type: "mouse_click",
							Data: string(data),
						}
					}
				}
			}

			time.Sleep(10 * time.Millisecond) // Polling interval
		}
	}
}

func (w *WindowsInputCapture) Close() error {
	return nil
}
