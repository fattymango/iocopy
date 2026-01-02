package screencapture

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"

	"github.com/kbinani/screenshot"
)

type ScreenCaptureer struct {
}

func NewScreenCapture() ScreenCapture {
	return &ScreenCaptureer{}
}

func (s *ScreenCaptureer) Capture() ([]byte, error) {
	img, err := s.captureScreen()
	if err != nil {
		log.Println("capture error:", err)
		return nil, err
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 60})
	if err != nil {
		log.Println("jpeg error:", err)
		return nil, err
	}

	data := buf.Bytes()
	return data, nil
}

func (s *ScreenCaptureer) captureScreen() (*image.RGBA, error) {
	// Find and capture the primary display
	// The primary display is typically the one with bounds starting at (0, 0)
	numDisplays := screenshot.NumActiveDisplays()
	
	var primaryBounds image.Rectangle
	
	// Find the display with bounds starting at (0, 0) - this is usually the primary
	for i := 0; i < numDisplays; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		if bounds.Min.X == 0 && bounds.Min.Y == 0 {
			primaryBounds = bounds
			break
		}
	}
	
	// If no display found at (0,0), use display 0 as fallback
	if primaryBounds.Empty() {
		primaryBounds = screenshot.GetDisplayBounds(0)
	}
	
	return screenshot.CaptureRect(primaryBounds)
}
