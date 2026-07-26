package procedural

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/render"
)

const (
	SpriteCount = 4
	spriteSize  = 16
	atlasW      = spriteSize * SpriteCount
	atlasH      = spriteSize
)

var _ render.AtlasSource = (*Atlas)(nil)

// Atlas is a ready-made render.AtlasSource: SpriteCount procedurally-drawn
// shapes (solid square, border, diamond, cross) — no image file required.
type Atlas struct {
	image *ebiten.Image
}

func NewAtlas() *Atlas {
	return &Atlas{image: buildAtlas()}
}

func (a *Atlas) Atlas() *ebiten.Image { return a.image }

func (a *Atlas) UV(spriteID uint8) (sx0, sy0, sx1, sy1 float32) {
	col := float32(spriteID % SpriteCount)
	sx0 = col * spriteSize
	sy0 = 0
	sx1 = sx0 + spriteSize
	sy1 = spriteSize
	return
}

func buildAtlas() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, atlasW, atlasH))
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	transparent := color.RGBA{}

	for id := 0; id < SpriteCount; id++ {
		ox := id * spriteSize
		switch id {
		case 0: // solid square
			for y := 0; y < spriteSize; y++ {
				for x := 0; x < spriteSize; x++ {
					img.SetRGBA(ox+x, y, white)
				}
			}
		case 1: // border only
			for y := 0; y < spriteSize; y++ {
				for x := 0; x < spriteSize; x++ {
					if x <= 1 || x >= spriteSize-2 || y <= 1 || y >= spriteSize-2 {
						img.SetRGBA(ox+x, y, white)
					} else {
						img.SetRGBA(ox+x, y, transparent)
					}
				}
			}
		case 2: // diamond
			cx, cy := spriteSize/2, spriteSize/2
			for y := 0; y < spriteSize; y++ {
				for x := 0; x < spriteSize; x++ {
					dist := abs(x-cx) + abs(y-cy)
					if dist <= cx-1 {
						img.SetRGBA(ox+x, y, white)
					} else {
						img.SetRGBA(ox+x, y, transparent)
					}
				}
			}
		case 3: // cross
			for y := 0; y < spriteSize; y++ {
				for x := 0; x < spriteSize; x++ {
					inH := y >= spriteSize/4 && y < spriteSize*3/4
					inV := x >= spriteSize/4 && x < spriteSize*3/4
					if inH || inV {
						img.SetRGBA(ox+x, y, white)
					} else {
						img.SetRGBA(ox+x, y, transparent)
					}
				}
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
