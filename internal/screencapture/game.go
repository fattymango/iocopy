package screencapture

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	once  sync.Once
	frame image.Image
	mu    sync.RWMutex
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Run() error {
	g.once.Do(func() {
		go func() {
			ebiten.SetWindowSize(1280, 800)
			ebiten.SetWindowTitle("Screen Stream Viewer")

			if err := ebiten.RunGame(g); err != nil {
				log.Fatal(err)
			}
		}()
	})

	return nil
}
func (g *Game) SetFrame(data []byte) error {
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("[game] Failed to decode jpeg: %v", err)
			return err
		}
		g.mu.Lock()
		g.frame = img
		g.mu.Unlock()
		return nil
	}
	if len(data) >= 8 {
		w := int(binary.BigEndian.Uint32(data[0:4]))
		h := int(binary.BigEndian.Uint32(data[4:8]))
		pix := data[8:]
		if w > 0 && h > 0 && len(pix) == w*h*4 {
			copyPix := make([]byte, len(pix))
			copy(copyPix, pix)
			g.mu.Lock()
			g.frame = &image.RGBA{Pix: copyPix, Stride: 4 * w, Rect: image.Rect(0, 0, w, h)}
			g.mu.Unlock()
			return nil
		}
	}
	return fmt.Errorf("unrecognized frame format (len=%d)", len(data))
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.mu.RLock()
	img := g.frame
	g.mu.RUnlock()

	if img == nil {
		return
	}

	eimg := ebiten.NewImageFromImage(img)

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(eimg, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
