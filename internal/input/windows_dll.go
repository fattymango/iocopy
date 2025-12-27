//go:build windows
// +build windows

package input

import "golang.org/x/sys/windows"

var (
	user32                  *windows.LazyDLL
	procGetAsyncKeyState    *windows.LazyProc
	procGetCursorPos        *windows.LazyProc
	procSetWindowsHookEx    *windows.LazyProc
	procUnhookWindowsHookEx *windows.LazyProc
	procCallNextHookEx      *windows.LazyProc
	procGetMessage          *windows.LazyProc
	procTranslateMessage    *windows.LazyProc
	procDispatchMessage     *windows.LazyProc
	procSendInput           *windows.LazyProc
	procSetCursorPos        *windows.LazyProc
	procmouse_event         *windows.LazyProc
	kernel32                *windows.LazyDLL
	procGetModuleHandle     *windows.LazyProc
)

func initWindowsDLLs() error {
	if user32 != nil {
		return nil // Already initialized
	}

	user32 = windows.NewLazyDLL("user32.dll")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
	procGetCursorPos = user32.NewProc("GetCursorPos")
	procSetWindowsHookEx = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx = user32.NewProc("CallNextHookEx")
	procGetMessage = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage = user32.NewProc("DispatchMessageW")
	procSendInput = user32.NewProc("SendInput")
	procSetCursorPos = user32.NewProc("SetCursorPos")
	procmouse_event = user32.NewProc("mouse_event")

	kernel32 = windows.NewLazyDLL("kernel32.dll")
	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")

	return nil
}

func getLastError() uint32 {
	ret, _, _ := kernel32.NewProc("GetLastError").Call()
	return uint32(ret)
}

