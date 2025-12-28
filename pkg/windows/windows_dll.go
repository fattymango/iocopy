//go:build windows
// +build windows

package windows

import "golang.org/x/sys/windows"

var (
	User32                         *windows.LazyDLL
	ProcGetAsyncKeyState           *windows.LazyProc
	ProcGetCursorPos               *windows.LazyProc
	ProcSetWindowsHookEx           *windows.LazyProc
	ProcUnhookWindowsHookEx        *windows.LazyProc
	ProcCallNextHookEx             *windows.LazyProc
	ProcGetMessage                 *windows.LazyProc
	ProcPeekMessage                *windows.LazyProc
	ProcTranslateMessage           *windows.LazyProc
	ProcDispatchMessage            *windows.LazyProc
	ProcSendInput                  *windows.LazyProc
	ProcSetCursorPos               *windows.LazyProc
	ProcMouseEvent                 *windows.LazyProc
	ProcCreateWindowEx             *windows.LazyProc
	ProcRegisterClassEx            *windows.LazyProc
	ProcRegisterClass              *windows.LazyProc
	ProcDefWindowProc              *windows.LazyProc
	ProcShowWindow                 *windows.LazyProc
	ProcSetWindowPos               *windows.LazyProc
	ProcDestroyWindow              *windows.LazyProc
	ProcGetSystemMetrics           *windows.LazyProc
	ProcPostQuitMessage            *windows.LazyProc
	ProcSetForegroundWindow        *windows.LazyProc
	ProcSetLayeredWindowAttributes *windows.LazyProc
	Gdi32                          *windows.LazyDLL
	ProcCreateSolidBrush           *windows.LazyProc
	Kernel32                       *windows.LazyDLL
	ProcGetModuleHandle            *windows.LazyProc
)

func InitWindowsDLLs() error {
	if User32 != nil {
		return nil // Already initialized
	}

	User32 = windows.NewLazyDLL("user32.dll")
	ProcGetAsyncKeyState = User32.NewProc("GetAsyncKeyState")
	ProcGetCursorPos = User32.NewProc("GetCursorPos")
	ProcSetWindowsHookEx = User32.NewProc("SetWindowsHookExW")
	ProcUnhookWindowsHookEx = User32.NewProc("UnhookWindowsHookEx")
	ProcCallNextHookEx = User32.NewProc("CallNextHookEx")
	ProcGetMessage = User32.NewProc("GetMessageW")
	ProcPeekMessage = User32.NewProc("PeekMessageW")
	ProcTranslateMessage = User32.NewProc("TranslateMessage")
	ProcDispatchMessage = User32.NewProc("DispatchMessageW")
	ProcSendInput = User32.NewProc("SendInput")
	ProcSetCursorPos = User32.NewProc("SetCursorPos")
	ProcMouseEvent = User32.NewProc("mouse_event")
	ProcCreateWindowEx = User32.NewProc("CreateWindowExW")
	ProcRegisterClassEx = User32.NewProc("RegisterClassExW")
	ProcRegisterClass = User32.NewProc("RegisterClassW")
	ProcDefWindowProc = User32.NewProc("DefWindowProcW")
	ProcShowWindow = User32.NewProc("ShowWindow")
	ProcSetWindowPos = User32.NewProc("SetWindowPos")
	ProcDestroyWindow = User32.NewProc("DestroyWindow")
	ProcGetSystemMetrics = User32.NewProc("GetSystemMetrics")
	ProcPostQuitMessage = User32.NewProc("PostQuitMessage")
	ProcSetForegroundWindow = User32.NewProc("SetForegroundWindow")
	ProcSetLayeredWindowAttributes = User32.NewProc("SetLayeredWindowAttributes")

	Gdi32 = windows.NewLazyDLL("gdi32.dll")
	ProcCreateSolidBrush = Gdi32.NewProc("CreateSolidBrush")

	Kernel32 = windows.NewLazyDLL("kernel32.dll")
	ProcGetModuleHandle = Kernel32.NewProc("GetModuleHandleW")

	return nil
}

func GetLastError() uint32 {
	ret, _, _ := Kernel32.NewProc("GetLastError").Call()
	return uint32(ret)
}
