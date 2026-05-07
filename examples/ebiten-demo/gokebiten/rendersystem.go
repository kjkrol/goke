package gokebiten

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
)

type RenderSystem interface {
	goke.System
	Draw(screen *ebiten.Image)
}
