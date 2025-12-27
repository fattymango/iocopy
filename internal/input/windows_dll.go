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
	procPeekMessage         *windows.LazyProc
	procTranslateMessage    *windows.LazyProc
	procDispatchMessage     *windows.LazyProc
	procSendInput           *windows.LazyProc
	procSetCursorPos        *windows.LazyProc
	procmouse_event         *windows.LazyProc
	procCreateWindowEx      *windows.LazyProc
	procRegisterClassEx     *windows.LazyProc
	procRegisterClass       *windows.LazyProc
	procDefWindowProc       *windows.LazyProc
	procShowWindow          *windows.LazyProc
	procSetWindowPos        *windows.LazyProc
	procDestroyWindow       *windows.LazyProc
	procGetSystemMetrics    *windows.LazyProc
	procPostQuitMessage     *windows.LazyProc
	procSetForegroundWindow      *windows.LazyProc
	procSetLayeredWindowAttributes *windows.LazyProc
	gdi32                        *windows.LazyDLL
	procCreateSolidBrush         *windows.LazyProc
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
	procPeekMessage = user32.NewProc("PeekMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage = user32.NewProc("DispatchMessageW")
	procSendInput = user32.NewProc("SendInput")
	procSetCursorPos = user32.NewProc("SetCursorPos")
	procmouse_event = user32.NewProc("mouse_event")
	procCreateWindowEx = user32.NewProc("CreateWindowExW")
	procRegisterClassEx = user32.NewProc("RegisterClassExW")
	procRegisterClass = user32.NewProc("RegisterClassW")
	procDefWindowProc = user32.NewProc("DefWindowProcW")
	procShowWindow = user32.NewProc("ShowWindow")
	procSetWindowPos = user32.NewProc("SetWindowPos")
	procDestroyWindow = user32.NewProc("DestroyWindow")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procPostQuitMessage = user32.NewProc("PostQuitMessage")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")

	gdi32 = windows.NewLazyDLL("gdi32.dll")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")

	kernel32 = windows.NewLazyDLL("kernel32.dll")
	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")

	return nil
}

func getLastError() uint32 {
	ret, _, _ := kernel32.NewProc("GetLastError").Call()
	return uint32(ret)
}

