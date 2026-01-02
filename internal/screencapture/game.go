package screencapture

import (
	"bytes"
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
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("[input] Failed to decode image: %v", err)
		return err
	}
	g.mu.Lock()
	g.frame = img
	g.mu.Unlock()
	return nil
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
