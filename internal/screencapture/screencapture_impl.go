package screencapture

import (
	"image"

	"github.com/kbinani/screenshot"
)

type ScreenCaptureer struct {
}

func NewScreenCapture() ScreenCapture {
	return &ScreenCaptureer{}
}

func (s *ScreenCaptureer) Bounds() (int, int) {
	bounds := screenshot.GetDisplayBounds(primaryDisplayIndex())
	return bounds.Dx(), bounds.Dy()
}

func (s *ScreenCaptureer) Capture() ([]byte, error) {
	img, err := s.captureScreen()
	if err != nil {
		return nil, err
	}
	pix := make([]byte, len(img.Pix))
	copy(pix, img.Pix)
	return pix, nil
}

func (s *ScreenCaptureer) captureScreen() (*image.RGBA, error) {
	bounds := screenshot.GetDisplayBounds(primaryDisplayIndex())
	return screenshot.CaptureRect(bounds)
}
