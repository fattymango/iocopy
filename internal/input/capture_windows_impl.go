package input

import (
	"encoding/json"
	"log"
	"runtime"
	"time"
	"unsafe"
)

// Windows DLL variables are declared in windows_dll.go (with build tags)
// This allows the code to compile on all platforms without build tags in this file

const (
	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_RBUTTONUP   = 0x0205
	WM_MBUTTONDOWN = 0x0207
	WM_MBUTTONUP   = 0x0208
	WM_MOUSEWHEEL  = 0x020A
	VK_CONTROL     = 0x11
	VK_SHIFT       = 0x10
	VK_B           = 0x42
	VK_LBUTTON     = 0x01
	VK_RBUTTON     = 0x02
	VK_MBUTTON     = 0x04
)

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
	if err := initWindowsDLLs(); err != nil {
		return nil, err
	}
	return &WindowsInputCapture{
		stopCh: make(chan struct{}),
	}, nil
}

func (w *WindowsInputCapture) Capture(eventCh chan<- InputEvent, stopCh <-chan struct{}) error {
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

func (w *WindowsInputCapture) captureKeyboard(eventCh chan<- InputEvent, stopCh <-chan struct{}) {
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
	keyNames[VK_CONTROL] = "Control_L"
	keyNames[VK_SHIFT] = "Shift_L"
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
				state, _, _ := procGetAsyncKeyState.Call(uintptr(vkCode))
				isPressed := (state & 0x8000) != 0

				wasPressed := keyStates[vkCode]
				if isPressed != wasPressed {
					keyStates[vkCode] = isPressed

					// Check modifiers (always check current state)
					ctrlState, _, _ := procGetAsyncKeyState.Call(uintptr(VK_CONTROL))
					shiftState, _, _ := procGetAsyncKeyState.Call(uintptr(VK_SHIFT))
					altState, _, _ := procGetAsyncKeyState.Call(uintptr(0x12)) // VK_MENU (Alt)

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
					if vkCode != VK_CONTROL && vkCode != VK_SHIFT && vkCode != 0x12 {
						kbEvent := KeyboardEvent{
							Key:       keyName,
							Action:    "press",
							Modifiers: modifiers,
						}
						if !isPressed {
							kbEvent.Action = "release"
						}

						data, _ := json.Marshal(kbEvent)
						eventCh <- InputEvent{
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

func (w *WindowsInputCapture) captureMouse(eventCh chan<- InputEvent, stopCh <-chan struct{}) {
	if runtime.GOOS != "windows" {
		return
	}

	var lastX, lastY int32
	buttonStates := make(map[uint32]bool) // Track button press states
	lastClickTime := make(map[uint32]int64) // Track last click time for double-click detection
	lastClickPos := make(map[uint32]POINT) // Track last click position
	inDoubleClickSequence := make(map[uint32]bool) // Track if we're in a double-click sequence

	// Button mappings
	buttons := map[uint32]string{
		VK_LBUTTON: "left",
		VK_RBUTTON: "right",
		VK_MBUTTON: "middle",
	}

	for {
		select {
		case <-stopCh:
			return
		default:
			var pt POINT
			procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

			// Check mouse movement
			if pt.X != lastX || pt.Y != lastY {
				moveEvent := MouseMoveEvent{
					X: int(pt.X),
					Y: int(pt.Y),
				}
				data, _ := json.Marshal(moveEvent)
				eventCh <- InputEvent{
					Type: "mouse_move",
					Data: string(data),
				}
				lastX, lastY = pt.X, pt.Y
			}

			// Check mouse buttons
			now := time.Now().UnixNano()
			doubleClickDelay := int64(500 * time.Millisecond) // Windows default double-click time (~500ms)
			maxDoubleClickDistance := int32(5) // Maximum pixels for double-click detection

			for vkCode, buttonName := range buttons {
				state, _, _ := procGetAsyncKeyState.Call(uintptr(vkCode))
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
							clickEvent := MouseClickEvent{
								Button:   buttonName,
								Action:   "double",
								X:        int(pt.X),
								Y:        int(pt.Y),
								IsDouble: true,
							}
							data, _ := json.Marshal(clickEvent)
							eventCh <- InputEvent{
								Type: "mouse_click",
								Data: string(data),
							}
						} else {
							// Normal single click press
							clickEvent := MouseClickEvent{
								Button:   buttonName,
								Action:   "press",
								X:        int(pt.X),
								Y:        int(pt.Y),
								IsDouble: false,
							}
							data, _ := json.Marshal(clickEvent)
							eventCh <- InputEvent{
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
						clickEvent := MouseClickEvent{
							Button:   buttonName,
							Action:   "release",
							X:        int(pt.X),
							Y:        int(pt.Y),
							IsDouble: false,
						}

						data, _ := json.Marshal(clickEvent)
						eventCh <- InputEvent{
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
