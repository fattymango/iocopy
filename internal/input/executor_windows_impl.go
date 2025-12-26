package input

import (
	"log"
	"runtime"
	"unsafe"
)

// WindowsInputExecutor executes input on Windows
type WindowsInputExecutor struct{}

func NewWindowsInputExecutor() (InputExecutor, error) {
	if err := initWindowsDLLs(); err != nil {
		return nil, err
	}
	return &WindowsInputExecutor{}, nil
}

type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	Mi   MOUSEINPUT
	Hi   HARDWAREINPUT
}

type KEYBDINPUT struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type HARDWAREINPUT struct {
	Msg    uint32
	ParamL uint16
	ParamH uint16
}

const (
	INPUT_KEYBOARD         = 1
	INPUT_MOUSE            = 0
	KEYEVENTF_KEYUP        = 0x0002
	MOUSEEVENTF_LEFTDOWN   = 0x0002
	MOUSEEVENTF_LEFTUP     = 0x0004
	MOUSEEVENTF_RIGHTDOWN  = 0x0008
	MOUSEEVENTF_RIGHTUP    = 0x0010
	MOUSEEVENTF_MIDDLEDOWN = 0x0020
	MOUSEEVENTF_MIDDLEUP   = 0x0040
	MOUSEEVENTF_WHEEL      = 0x0800
	MOUSEEVENTF_ABSOLUTE   = 0x8000
)

var keyMap = map[string]uint16{
	"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44, "e": 0x45, "f": 0x46,
	"g": 0x47, "h": 0x48, "i": 0x49, "j": 0x4A, "k": 0x4B, "l": 0x4C,
	"m": 0x4D, "n": 0x4E, "o": 0x4F, "p": 0x50, "q": 0x51, "r": 0x52,
	"s": 0x53, "t": 0x54, "u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58,
	"y": 0x59, "z": 0x5A,
	"Return": 0x0D, "Escape": 0x1B, "Tab": 0x09, "space": 0x20,
	"BackSpace": 0x08, "Control_L": VK_CONTROL, "Shift_L": VK_SHIFT,
}

func (w *WindowsInputExecutor) ExecuteKeyboard(event KeyboardEvent) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	vkCode, ok := keyMap[event.Key]
	if !ok {
		// Try to parse as hex or use default
		log.Printf("[input] Unknown key: %s", event.Key)
		return nil
	}

	var inputs [1]INPUT
	inputs[0].Type = INPUT_KEYBOARD
	inputs[0].Ki.Vk = vkCode
	inputs[0].Ki.Flags = 0
	if event.Action == "release" {
		inputs[0].Ki.Flags = KEYEVENTF_KEYUP
	}

	// Handle modifiers
	if contains(event.Modifiers, "ctrl") {
		// Send Ctrl
		var ctrlInput INPUT
		ctrlInput.Type = INPUT_KEYBOARD
		ctrlInput.Ki.Vk = VK_CONTROL
		if event.Action == "release" {
			ctrlInput.Ki.Flags = KEYEVENTF_KEYUP
		}
		procSendInput.Call(1, uintptr(unsafe.Pointer(&ctrlInput)), unsafe.Sizeof(INPUT{}))
	}
	if contains(event.Modifiers, "shift") {
		// Send Shift
		var shiftInput INPUT
		shiftInput.Type = INPUT_KEYBOARD
		shiftInput.Ki.Vk = VK_SHIFT
		if event.Action == "release" {
			shiftInput.Ki.Flags = KEYEVENTF_KEYUP
		}
		procSendInput.Call(1, uintptr(unsafe.Pointer(&shiftInput)), unsafe.Sizeof(INPUT{}))
	}

	procSendInput.Call(1, uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(INPUT{}))
	return nil
}

func (w *WindowsInputExecutor) ExecuteMouseMove(event MouseMoveEvent) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	procSetCursorPos.Call(uintptr(event.X), uintptr(event.Y))
	return nil
}

func (w *WindowsInputExecutor) ExecuteMouseClick(event MouseClickEvent) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	var flags uint32
	switch event.Button {
	case "left":
		if event.Action == "press" {
			flags = MOUSEEVENTF_LEFTDOWN
		} else {
			flags = MOUSEEVENTF_LEFTUP
		}
	case "right":
		if event.Action == "press" {
			flags = MOUSEEVENTF_RIGHTDOWN
		} else {
			flags = MOUSEEVENTF_RIGHTUP
		}
	case "middle":
		if event.Action == "press" {
			flags = MOUSEEVENTF_MIDDLEDOWN
		} else {
			flags = MOUSEEVENTF_MIDDLEUP
		}
	}

	procSetCursorPos.Call(uintptr(event.X), uintptr(event.Y))
	procmouse_event.Call(uintptr(flags), 0, 0, 0, 0)
	return nil
}

func (w *WindowsInputExecutor) ExecuteMouseScroll(event MouseScrollEvent) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	delta := uint32(event.DeltaY * 120) // WHEEL_DELTA = 120
	procmouse_event.Call(uintptr(MOUSEEVENTF_WHEEL), 0, 0, uintptr(delta), 0)
	return nil
}

func (w *WindowsInputExecutor) Close() error {
	return nil
}
