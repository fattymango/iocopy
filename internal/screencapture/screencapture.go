package screencapture

type ScreenCapture interface {
	Capture() ([]byte, error)
}
