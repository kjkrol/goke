package main

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
)

type EntitiesRendererSystem struct {
	*Resources
	renderView       *goke.View3[Position, Collision, Appearance]
	pixelImage       *ebiten.Image
	drawImageOptions *ebiten.DrawImageOptions
}

var _ gokebiten.RenderSystem = (*EntitiesRendererSystem)(nil)

func NewEntitiesRendererSystem(resources *Resources) gokebiten.RenderSystem {
	pixelImage := ebiten.NewImage(resources.rectSize, resources.rectSize)
	pixelImage.Fill(color.White)
	return &EntitiesRendererSystem{
		Resources:        resources,
		pixelImage:       pixelImage,
		drawImageOptions: &ebiten.DrawImageOptions{},
	}
}

func (s *EntitiesRendererSystem) Init(ecs *goke.ECS) {
	s.renderView = goke.NewView3[Position, Collision, Appearance](ecs)
}

func (s *EntitiesRendererSystem) Update(lookup goke.Lookup, _ *goke.Schedule, d time.Duration) {}

func (s *EntitiesRendererSystem) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 50, G: 50, B: 50, A: 255})
	imgW := s.pixelImage.Bounds().Dx()
	imgH := s.pixelImage.Bounds().Dy()
	now := time.Now()
	for page := range s.renderView.All() {
		for i, _ := range page.Entity {
			aabb, col, app := &page.Comp1[i], &page.Comp2[i], &page.Comp3[i]
			s.draw(aabb.AABB.AABB, app, col, screen, imgW, imgH, now)
			aabb.AABB.VisitFragments(func(pos plane.FragPosition, box geom.AABB[uint32]) bool {
				s.draw(box, app, col, screen, imgW, imgH, now)
				return true
			})
		}
	}
}

func (s *EntitiesRendererSystem) draw(
	aabb geom.AABB[uint32],
	app *Appearance,
	col *Collision,
	screen *ebiten.Image,
	imgW, imgH int,
	now time.Time,
) {
	op := s.drawImageOptions
	// Resetujemy opcje (geoM - macierz transformacji)
	op.GeoM.Reset()

	// Skalowanie (width, height)
	w := float64(aabb.BottomRight.X - aabb.TopLeft.X)
	h := float64(aabb.BottomRight.Y - aabb.TopLeft.Y)
	op.GeoM.Scale(w/float64(imgW), h/float64(imgH))

	// Przesunięcie (pozycja X, Y)
	op.GeoM.Translate(float64(aabb.TopLeft.X), float64(aabb.TopLeft.Y))

	// Kolorowanie
	op.ColorScale.Reset()
	if now.Sub(col.timestamp) < 40*time.Millisecond {
		op.ColorScale.ScaleWithColor(color.RGBA{R: 255, A: 255})
	} else {
		op.ColorScale.ScaleWithColor(app.Color)
	}
	screen.DrawImage(s.pixelImage, op)
}
