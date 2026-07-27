package render

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke/v2"
)

// Renderer draws each frame; Init runs once at registration, outside the
// per-tick Update cycle — there is no scheduled Update here.
type Renderer interface {
	Init(*goke.ECS)
	Draw(screen *ebiten.Image)
}
