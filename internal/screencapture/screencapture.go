package screencapture

type ScreenCapture interface {
	// Bounds returns the primary display size in pixels (width × height).
	Bounds() (width, height int)
	// Capture returns raw RGBA pixels, row-major, stride = 4 * width (same layout as image.RGBA.Pix).
	Capture() ([]byte, error)
}
