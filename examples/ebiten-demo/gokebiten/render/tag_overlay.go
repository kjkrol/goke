package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
)

var _ Renderer = (*TagOverlayRenderer)(nil)

// TagOverlayRenderer draws sprite spriteID from atlas, tinted c, over every
// entity matched by filter (e.g. goke.Include[SomeTag]()) — reusable for
// any tag, not just one specific effect.
type TagOverlayRenderer struct {
	query    *goke.Query
	pos      goke.Comp[kinematics.Position]
	spriteID uint8
	color    color.RGBA
	filter   goke.Opt
	batch    spriteBatch
}

func NewTagOverlayRenderer(atlas AtlasSource, camera *Camera, spriteID uint8, c color.RGBA, filter goke.Opt) *TagOverlayRenderer {
	return &TagOverlayRenderer{
		batch:    newSpriteBatch(atlas, camera),
		spriteID: spriteID,
		color:    c,
		filter:   filter,
	}
}

func (r *TagOverlayRenderer) Init(si *goke.SysInit) {
	r.query = si.NewQueryBuilder(&r.pos).Include(r.filter).Build()
}

func (r *TagOverlayRenderer) Draw(screen *ebiten.Image) {
	r.batch.reset()

	r.query.All()
	for r.query.Next() {
		cursor := r.query.Cursor()
		positions := r.pos.Slice(cursor)
		for i := range cursor.IDs {
			r.batch.drawQuad(positions[i], r.spriteID, r.color)
		}
	}

	r.batch.flush(screen)
}
