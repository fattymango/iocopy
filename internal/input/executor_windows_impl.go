package input

import (
	"log"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WindowsInputExecutor executes input on Windows
type WindowsInputExecutor struct{}

func NewWindowsInputExecutor() (InputExecutor, error) {
	if err := initWindowsDLLs(); err != nil {
		return nil, err
	}
	return &WindowsInputExecutor{}, nil
}

// INPUT structure for Windows SendInput
// Windows INPUT is 40 bytes on 64-bit: Type(4) + padding(4) + union(32)
// The union must be aligned to 8-byte boundary, so we add padding before KEYBDINPUT
type INPUT struct {
	Type uint32
	_    [4]byte    // Padding to align union to 8-byte boundary
	Ki   KEYBDINPUT // KEYBDINPUT starts at offset 8
	_    [8]byte    // Padding to make total size 40 bytes (KEYBDINPUT is 20 bytes, so 8 more needed)
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
	VK_MENU                = 0x12 // Alt key
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

	// Skip modifier keys themselves - they're handled separately
	if event.Key == "Control_L" || event.Key == "Shift_L" {
		return nil
	}

	vkCode, ok := keyMap[event.Key]
	if !ok {
		log.Printf("[input] Unknown key: %s", event.Key)
		return nil
	}

	log.Printf("[input] Executing keyboard event: Key=%s, Action=%s, Modifiers=%v", event.Key, event.Action, event.Modifiers)

	if event.Action == "press" {
		// For press: send modifiers down first, then the key
		var inputs []INPUT

		// Press modifiers first
		if contains(event.Modifiers, "ctrl") {
			ctrlInput := INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					Vk:    VK_CONTROL,
					Flags: 0,
				},
			}
			inputs = append(inputs, ctrlInput)
		}
		if contains(event.Modifiers, "shift") {
			shiftInput := INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					Vk:    VK_SHIFT,
					Flags: 0,
				},
			}
			inputs = append(inputs, shiftInput)
		}
		if contains(event.Modifiers, "alt") {
			altInput := INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					Vk:    VK_MENU,
					Flags: 0,
				},
			}
			inputs = append(inputs, altInput)
		}

		// Then press the key
		keyInput := INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				Vk:    vkCode,
				Flags: 0,
			},
		}
		inputs = append(inputs, keyInput)

		// Send all inputs at once
		if len(inputs) > 0 {
			ret, _, err := procSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(INPUT{}))
			if ret == 0 {
				// Get the actual Windows error code
				kernel32 := windows.NewLazyDLL("kernel32.dll")
				procGetLastError := kernel32.NewProc("GetLastError")
				errCode, _, _ := procGetLastError.Call()
				log.Printf("[input] SendInput failed for key press: error code %d, errno %v, input size %d, inputs count %d", errCode, err, unsafe.Sizeof(INPUT{}), len(inputs))
			} else {
				log.Printf("[input] SendInput succeeded: sent %d inputs", ret)
			}
		}
	} else if event.Action == "release" {
		// For release: release the key first, then the modifiers
		var inputs []INPUT

		// Release the key first
		keyInput := INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				Vk:    vkCode,
				Flags: KEYEVENTF_KEYUP,
			},
		}
		inputs = append(inputs, keyInput)

		// Then release modifiers (in reverse order)
		if contains(event.Modifiers, "alt") {
			altInput := INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					Vk:    VK_MENU,
					Flags: KEYEVENTF_KEYUP,
				},
			}
			inputs = append(inputs, altInput)
		}
		if contains(event.Modifiers, "shift") {
			shiftInput := INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					Vk:    VK_SHIFT,
					Flags: KEYEVENTF_KEYUP,
				},
			}
			inputs = append(inputs, shiftInput)
		}
		if contains(event.Modifiers, "ctrl") {
			ctrlInput := INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					Vk:    VK_CONTROL,
					Flags: KEYEVENTF_KEYUP,
				},
			}
			inputs = append(inputs, ctrlInput)
		}

		// Send all inputs at once
		if len(inputs) > 0 {
			ret, _, err := procSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(INPUT{}))
			if ret == 0 {
				// Get the actual Windows error code
				// Get the actual Windows error code
				kernel32 := windows.NewLazyDLL("kernel32.dll")
				procGetLastError := kernel32.NewProc("GetLastError")
				errCode, _, _ := procGetLastError.Call()
				log.Printf("[input] SendInput failed for key release: error code %d, errno %v, input size %d, inputs count %d", errCode, err, unsafe.Sizeof(INPUT{}), len(inputs))
			} else {
				log.Printf("[input] SendInput succeeded: sent %d inputs", ret)
			}
		}
	}

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

	log.Printf("[input] Executing mouse click: Button=%s, Action=%s, X=%d, Y=%d, IsDouble=%v", event.Button, event.Action, event.X, event.Y, event.IsDouble)

	// Set cursor position first
	procSetCursorPos.Call(uintptr(event.X), uintptr(event.Y))

	// Determine button flags
	var downFlag, upFlag uint32
	switch event.Button {
	case "left":
		downFlag = MOUSEEVENTF_LEFTDOWN
		upFlag = MOUSEEVENTF_LEFTUP
	case "right":
		downFlag = MOUSEEVENTF_RIGHTDOWN
		upFlag = MOUSEEVENTF_RIGHTUP
	case "middle":
		downFlag = MOUSEEVENTF_MIDDLEDOWN
		upFlag = MOUSEEVENTF_MIDDLEUP
	default:
		log.Printf("[input] Unknown mouse button: %s", event.Button)
		return nil
	}

	// Use INPUT structure with MOUSEINPUT data
	// We'll overlay MOUSEINPUT into the union area (same 40-byte structure)

	// Create mouse input structure matching Windows INPUT (40 bytes)
	// Use unsafe to overlay MOUSEINPUT into the union area
	type MOUSE_INPUT_RAW struct {
		Type uint32
		_    [4]byte  // Padding for union alignment (8-byte boundary)
		Data [32]byte // Union data area - we'll put MOUSEINPUT here
	}

	createMouseInput := func(flags uint32) MOUSE_INPUT_RAW {
		var input MOUSE_INPUT_RAW
		input.Type = INPUT_MOUSE
		mi := MOUSEINPUT{
			Flags: flags,
			// Dx, Dy are 0 for relative movement (we use SetCursorPos instead)
			// MouseData is 0 for button clicks
		}
		// Copy MOUSEINPUT into the Data area
		*(*MOUSEINPUT)(unsafe.Pointer(&input.Data[0])) = mi
		return input
	}

	if event.Action == "double" || event.IsDouble {
		// Double click: send press, release, press, release in quick succession
		var inputs []MOUSE_INPUT_RAW

		// First click: press and release
		inputs = append(inputs, createMouseInput(downFlag))
		inputs = append(inputs, createMouseInput(upFlag))
		// Second click: press and release (Windows double-click timing is handled automatically)
		inputs = append(inputs, createMouseInput(downFlag))
		inputs = append(inputs, createMouseInput(upFlag))

		// Send all inputs
		if len(inputs) > 0 {
			ret, _, err := procSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(MOUSE_INPUT_RAW{}))
			if ret == 0 {
				kernel32 := windows.NewLazyDLL("kernel32.dll")
				procGetLastError := kernel32.NewProc("GetLastError")
				errCode, _, _ := procGetLastError.Call()
				log.Printf("[input] SendInput failed for mouse double click: error code %d, errno %v", errCode, err)
			} else {
				log.Printf("[input] SendInput succeeded for double click: sent %d inputs", ret)
			}
		}
	} else if event.Action == "press" {
		// Single press
		input := createMouseInput(downFlag)

		ret, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(MOUSE_INPUT_RAW{}))
		if ret == 0 {
			kernel32 := windows.NewLazyDLL("kernel32.dll")
			procGetLastError := kernel32.NewProc("GetLastError")
			errCode, _, _ := procGetLastError.Call()
			log.Printf("[input] SendInput failed for mouse press: error code %d, errno %v", errCode, err)
		}
	} else if event.Action == "release" {
		// Single release
		input := createMouseInput(upFlag)

		ret, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(MOUSE_INPUT_RAW{}))
		if ret == 0 {
			kernel32 := windows.NewLazyDLL("kernel32.dll")
			procGetLastError := kernel32.NewProc("GetLastError")
			errCode, _, _ := procGetLastError.Call()
			log.Printf("[input] SendInput failed for mouse release: error code %d, errno %v", errCode, err)
		}
	}

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
