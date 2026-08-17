package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
)

var _ Renderer = (*EntitiesRenderer)(nil)

// EntitiesRenderer draws every entity with Position+Appearance as a
// toroidal-fragment-aware sprite quad, batched into a single DrawTriangles
// call. It knows nothing about why an Appearance has the color/sprite it
// has — that's the game's job (see spatial.EntityExtras); layer other
// Renderers (e.g. TagOverlayRenderer) on top for status-driven visuals.
type EntitiesRenderer struct {
	renderQuery *goke.Query
	pos         goke.Comp[kinematics.Position]
	appearance  goke.Comp[Appearance]
	exclude     []goke.Opt
	batch       spriteBatch
}

// NewEntitiesRenderer draws every entity with Position+Appearance, except
// any matching an exclude filter (e.g. goke.Exclude[SomeTag]()) — for tags
// a different Renderer (e.g. TagOverlayRenderer) already draws in full.
func NewEntitiesRenderer(atlas AtlasSource, camera *Camera, exclude ...goke.Opt) *EntitiesRenderer {
	return &EntitiesRenderer{batch: newSpriteBatch(atlas, camera), exclude: exclude}
}

func (s *EntitiesRenderer) Init(si *goke.SysInit) {
	b := si.NewQueryBuilder(&s.pos, &s.appearance)
	if len(s.exclude) > 0 {
		b = b.Exclude(s.exclude...)
	}
	s.renderQuery = b.Build()
}

func (s *EntitiesRenderer) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 50, G: 50, B: 50, A: 255})
	s.batch.reset()

	s.renderQuery.All()
	for s.renderQuery.Next() {
		cursor := s.renderQuery.Cursor()
		positions := s.pos.Slice(cursor)
		appearances := s.appearance.Slice(cursor)
		for i := range cursor.IDs {
			s.batch.drawQuad(positions[i], appearances[i].SpriteID, appearances[i].Color)
		}
	}

	s.batch.flush(screen)
}
