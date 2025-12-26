//go:build !windows
// +build !windows

package input

// Stub variables for non-Windows platforms
// These will never be used but allow the code to compile
var (
	procGetAsyncKeyState    interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procGetCursorPos        interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procSetWindowsHookEx    interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procUnhookWindowsHookEx interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procCallNextHookEx      interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procGetMessage          interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procTranslateMessage    interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procDispatchMessage     interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procSendInput           interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procSetCursorPos        interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procmouse_event         interface{ Call(...uintptr) (uintptr, uintptr, error) }
	procGetModuleHandle     interface{ Call(...uintptr) (uintptr, uintptr, error) }
)

func initWindowsDLLs() error {
	return nil // Not on Windows, no initialization needed
}

