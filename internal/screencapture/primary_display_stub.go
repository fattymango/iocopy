//go:build !windows

package screencapture

func primaryDisplayIndex() int {
	return 0
}
