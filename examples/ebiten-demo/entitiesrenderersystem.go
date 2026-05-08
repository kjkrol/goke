package main

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/plane"
)

type EntitiesRendererSystem struct {
	*Resource
	renderView       *goke.View2[Position, Appearance]
	pixelImage       *ebiten.Image
	drawImageOptions *ebiten.DrawImageOptions
}

var _ gokebiten.RenderSystem = (*EntitiesRendererSystem)(nil)

func NewEntitiesRendererSystem(resource *Resource) *EntitiesRendererSystem {
	pixelImage := ebiten.NewImage(resource.rectSize, resource.rectSize)
	pixelImage.Fill(color.White)
	return &EntitiesRendererSystem{
		Resource:         resource,
		pixelImage:       pixelImage,
		drawImageOptions: &ebiten.DrawImageOptions{},
	}
}

func (s *EntitiesRendererSystem) Init(ecs *goke.ECS) {
	s.renderView = goke.NewView2[Position, Appearance](ecs)
}

func (s *EntitiesRendererSystem) Update(lookup goke.Lookup, _ *goke.Schedule, d time.Duration) {}

func (s *EntitiesRendererSystem) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{R: 50, G: 50, B: 50, A: 255})

	// Opcje rysowania, alokujemy raz, żeby nie śmiecić pamięci

	imgW, imgH := s.pixelImage.Size()
	for head := range s.renderView.Values() {
		aabb, app := head.V1, head.V2
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
		op.ColorScale.ScaleWithColor(app.Color)

		// To jest BARDZO szybkie (batching na GPU)
		screen.DrawImage(s.pixelImage, op)

		// Uwaga: Obsługę "fragmentów" torusa pomijam dla czytelności,
		// ale analogicznie używasz DrawImage zamiast FillRect.
		aabb.AABB.VisitFragments(func(pos plane.FragPosition, box geom.AABB[uint32]) bool {
			return true
		})
	}
}
