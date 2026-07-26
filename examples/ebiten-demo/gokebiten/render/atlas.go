package render

import "github.com/hajimehoshi/ebiten/v2"

// AtlasSource supplies the sprite sheet and per-sprite UV rects that
// spriteBatch draws quads from — swappable, e.g. procedurally-drawn shapes
// vs. sprites loaded from a file.
type AtlasSource interface {
	Atlas() *ebiten.Image
	UV(spriteID uint8) (sx0, sy0, sx1, sy1 float32)
}
