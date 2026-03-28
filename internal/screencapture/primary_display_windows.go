//go:build windows

package screencapture

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
)

func primaryDisplayIndex() int {
	var idx int
	found := -1
	cb := syscall.NewCallback(func(hMonitor, _, _, _ uintptr) uintptr {
		var info win.MONITORINFO
		info.CbSize = uint32(unsafe.Sizeof(info))
		if !win.GetMonitorInfo(win.HMONITOR(hMonitor), &info) {
			idx++
			return 1
		}
		if info.DwFlags&win.MONITORINFOF_PRIMARY != 0 {
			found = idx
			return 0
		}
		idx++
		return 1
	})
	procEnumDisplayMonitors.Call(0, 0, cb, 0)
	if found >= 0 {
		return found
	}
	return 0
}
