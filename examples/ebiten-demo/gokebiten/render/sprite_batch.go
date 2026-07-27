package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
)

const (
	fragScreenLeft    = plane.FRAG_RIGHT
	fragScreenTop     = plane.FRAG_BOTTOM
	fragScreenTopLeft = plane.FRAG_BOTTOM_RIGHT
)

// spriteBatch batches textured, toroidal-fragment-aware quads from an
// AtlasSource into a single DrawTriangles call, transformed through a
// Camera — shared by EntitiesRenderer and TagOverlayRenderer so both get
// correct world-wrap and camera handling without duplicating the quad math.
type spriteBatch struct {
	atlas    AtlasSource
	camera   *Camera
	vertices []ebiten.Vertex
	indices  []uint16
	triOpts  *ebiten.DrawTrianglesOptions
}

func newSpriteBatch(atlas AtlasSource, camera *Camera) spriteBatch {
	return spriteBatch{
		atlas:   atlas,
		camera:  camera,
		triOpts: &ebiten.DrawTrianglesOptions{FillRule: ebiten.FillAll},
	}
}

func (b *spriteBatch) reset() {
	b.vertices = b.vertices[:0]
	b.indices = b.indices[:0]
}

func (b *spriteBatch) drawQuad(pos kinematics.Position, spriteID uint8, c color.RGBA) {
	if !b.camera.Visible(pos.AABB.AABB) {
		return
	}

	sx0, sy0, sx1, sy1 := b.atlas.UV(spriteID)
	r, g, bl, a := colorToFloats(c)

	mainBox := pos.AABB.AABB
	hasFragments := false
	sizeX := float32(pos.AABB.Size.X)
	sizeY := float32(pos.AABB.Size.Y)

	pos.AABB.VisitFragments(func(fp plane.FragPosition, fragBox geom.AABB[uint32]) bool {
		hasFragments = true
		tlx := float32(fragBox.TopLeft.X)
		tly := float32(fragBox.TopLeft.Y)
		brx := float32(fragBox.BottomRight.X)
		bry := float32(fragBox.BottomRight.Y)

		switch fp {
		case fragScreenLeft:
			tlx = brx - sizeX
			bry = tly + sizeY
		case fragScreenTop:
			tly = bry - sizeY
			brx = tlx + sizeX
		case fragScreenTopLeft:
			tlx = brx - sizeX
			tly = bry - sizeY
		default:
			return true
		}
		b.appendFullQuad(tlx, tly, brx, bry, sx0, sy0, sx1, sy1, r, g, bl, a)
		return true
	})

	brx := float32(mainBox.BottomRight.X)
	bry := float32(mainBox.BottomRight.Y)
	tlx := float32(mainBox.TopLeft.X)
	tly := float32(mainBox.TopLeft.Y)

	if hasFragments {
		brx = tlx + sizeX
		bry = tly + sizeY
	}

	b.appendFullQuad(tlx, tly, brx, bry, sx0, sy0, sx1, sy1, r, g, bl, a)
}

func (b *spriteBatch) appendFullQuad(
	x0, y0, x1, y1 float32,
	sx0, sy0, sx1, sy1 float32,
	r, g, bl, a float32,
) {
	x0, y0 = b.camera.ToScreen(x0, y0)
	x1, y1 = b.camera.ToScreen(x1, y1)

	idx := uint16(len(b.vertices))
	b.vertices = append(b.vertices,
		ebiten.Vertex{DstX: x0, DstY: y0, SrcX: sx0, SrcY: sy0, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
		ebiten.Vertex{DstX: x1, DstY: y0, SrcX: sx1, SrcY: sy0, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
		ebiten.Vertex{DstX: x0, DstY: y1, SrcX: sx0, SrcY: sy1, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
		ebiten.Vertex{DstX: x1, DstY: y1, SrcX: sx1, SrcY: sy1, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
	)
	b.indices = append(b.indices, idx, idx+1, idx+2, idx+1, idx+2, idx+3)
}

func (b *spriteBatch) flush(screen *ebiten.Image) {
	if len(b.indices) > 0 {
		screen.DrawTriangles(b.vertices, b.indices, b.atlas.Atlas(), b.triOpts)
	}
}

func colorToFloats(c color.RGBA) (r, g, b, a float32) {
	return float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255, float32(c.A) / 255
}
