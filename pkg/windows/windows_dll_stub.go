//go:build !windows
// +build !windows

package windows

import "golang.org/x/sys/windows"

// Stub variables for non-Windows platforms
// These will never be used but allow the code to compile
var (
	ProcGetAsyncKeyState interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcGetCursorPos interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcSetWindowsHookEx interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcUnhookWindowsHookEx interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcCallNextHookEx interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcGetMessage interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcTranslateMessage interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcDispatchMessage interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcSendInput interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcSetCursorPos interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcMouseEvent interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcGetModuleHandle interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	ProcGetLastError interface {
		Call(...uintptr) (uintptr, uintptr, error)
	}
	Kernel32 interface {
		NewProc(string) *windows.LazyProc
	}
)

func InitWindowsDLLs() error {
	return nil // Not on Windows, no initialization needed
}
