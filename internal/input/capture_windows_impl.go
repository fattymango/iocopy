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

	for {
		select {
		case <-stopCh:
			return
		default:
			var pt POINT
			procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

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
			// Left button = 0x01, Right = 0x02, Middle = 0x10
			// This is simplified - full implementation would use hooks

			time.Sleep(16 * time.Millisecond) // ~60fps polling
		}
	}
}

func (w *WindowsInputCapture) Close() error {
	return nil
}
