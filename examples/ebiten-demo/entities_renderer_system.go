package main

import (
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
)

const (
	spriteSize = 16
	atlasW     = spriteSize * SpriteCount
	atlasH     = spriteSize
)

type EntitiesRendererSystem struct {
	*Resources
	renderView *goke.View3[Position, Collision, Appearance]
	atlas      *ebiten.Image
	vertices   []ebiten.Vertex
	indices    []uint16
	triOpts    *ebiten.DrawTrianglesOptions
}

var _ gokebiten.RenderSystem = (*EntitiesRendererSystem)(nil)

func NewEntitiesRendererSystem(resources *Resources) gokebiten.RenderSystem {
	return &EntitiesRendererSystem{
		Resources: resources,
		atlas:     buildAtlas(),
		triOpts:   &ebiten.DrawTrianglesOptions{FillRule: ebiten.FillAll},
	}
}

func (s *EntitiesRendererSystem) Init(ecs *goke.ECS) {
	s.renderView = goke.NewView3[Position, Collision, Appearance](ecs)
}

func (s *EntitiesRendererSystem) Update(_ goke.Lookup, _ *goke.Schedule, _ time.Duration) {}

const (
	FRAG_SCREEN_LEFT     = plane.FRAG_RIGHT
	FRAG_SCREEN_TOP      = plane.FRAG_BOTTOM
	FRAG_SCREEN_TOP_LEFT = plane.FRAG_BOTTOM_RIGHT
)

func (s *EntitiesRendererSystem) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 50, G: 50, B: 50, A: 255})
	now := time.Now()

	s.vertices = s.vertices[:0]
	s.indices = s.indices[:0]

	for page := range s.renderView.All() {
		for i := range page.Entity {
			pos, col, app := page.Comp1[i], page.Comp2[i], page.Comp3[i]
			sx0, sy0, sx1, sy1 := spriteUV(app.SpriteID)
			r, g, b, a := s.resolveColor(app, col, now)

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
				case FRAG_SCREEN_LEFT:
					tlx = brx - sizeX
					bry = tly + sizeY
				case FRAG_SCREEN_TOP:
					tly = bry - sizeY
					brx = tlx + sizeX
				case FRAG_SCREEN_TOP_LEFT:
					tlx = brx - sizeX
					tly = bry - sizeY
				default:
					return true
				}
				s.appendFullQuad(tlx, tly, brx, bry, sx0, sy0, sx1, sy1, r, g, b, a)
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

			s.appendFullQuad(tlx, tly, brx, bry, sx0, sy0, sx1, sy1, r, g, b, a)

		}
	}

	if len(s.indices) > 0 {
		screen.DrawTriangles(s.vertices, s.indices, s.atlas, s.triOpts)
	}
}

func (s *EntitiesRendererSystem) resolveColor(app Appearance, col Collision, now time.Time) (r, g, b, a float32) {
	if now.Sub(col.timestamp) < 40*time.Millisecond {
		return 1, 0, 0, 1
	}
	return float32(app.Color.R) / 255,
		float32(app.Color.G) / 255,
		float32(app.Color.B) / 255,
		float32(app.Color.A) / 255
}

func (s *EntitiesRendererSystem) appendFullQuad(
	x0, y0, x1, y1 float32,
	sx0, sy0, sx1, sy1 float32,
	r, g, b, a float32,
) {
	idx := uint16(len(s.vertices))
	s.vertices = append(s.vertices,
		ebiten.Vertex{DstX: x0, DstY: y0, SrcX: sx0, SrcY: sy0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		ebiten.Vertex{DstX: x1, DstY: y0, SrcX: sx1, SrcY: sy0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		ebiten.Vertex{DstX: x0, DstY: y1, SrcX: sx0, SrcY: sy1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		ebiten.Vertex{DstX: x1, DstY: y1, SrcX: sx1, SrcY: sy1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	)
	s.indices = append(s.indices, idx, idx+1, idx+2, idx+1, idx+2, idx+3)
}

func spriteUV(id uint8) (sx0, sy0, sx1, sy1 float32) {
	col := float32(id % SpriteCount)
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

	atlas := ebiten.NewImageFromImage(img)
	return atlas
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
