package main

import (
	"image/color"
	"math"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/control"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions/strategies/elastic"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions/strategies/stats"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics/spawners/grid"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics/spawners/randomvelocity"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/render"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/render/atlases/procedural"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/spatial"
	"github.com/kjkrol/goke/v2"
)

const (
	TPS          = 60 * 2
	ScreenWidth  = 1024
	ScreenHeight = 1024
	RectSize     = 20
	FillPercent  = 20
)

var EntityCount = int(math.Floor(FillPercent / 100.0 * float64(ScreenWidth*ScreenHeight) / float64(RectSize*RectSize)))

type State struct{}

type Telemetry struct {
	Kinematics kinematics.Telemetry
	Collision  stats.Stats
}

func (t *Telemetry) Reset() { t.Collision.Counter = 0 }

var _ control.EventHandler = (*EventHandler)(nil)

type EventHandler struct{ game *gokebiten.Game }

func (e *EventHandler) HandleEvents(events *control.InputEvents) {
	for _, k := range events.KeyEvents {
		if k.Key == ebiten.KeySpace && k.Action == control.ActionPress {
			e.game.TogglePause()
		}
	}
}

func main() {
	resources := gokebiten.NewResources(
		&gokebiten.GameProps{
			Title:       "GOKe + GOKg + Ebiten Integration",
			ScreenWidth: ScreenWidth, ScreenHeight: ScreenHeight,
			TargetTPS: TPS},
		spatial.Config{Width: ScreenWidth, Height: ScreenHeight, Toroidal: true},
		State{},
		Telemetry{},
	)

	game := gokebiten.NewGame(resources)
	game.SetEventHandler(&EventHandler{game: game})

	game.UseWorld(spatial.Population{MaxCount: EntityCount, MinSize: RectSize, MaxSize: RectSize})

	physics := game.UsePhysics()
	physics.SetCollisionHandlers(
		elastic.NewHandler(),
		stats.NewHandler(&resources.Telemetry().Collision),
	)
	physics.SetHitExpires(100 * time.Millisecond)

	game.Loop(func(ctx goke.RunCtx, d time.Duration) {
		physics.Run(ctx, d)
		ctx.Sync()
	})

	camera := render.NewCamera(ScreenWidth, ScreenHeight)
	atlas := procedural.NewAtlas()

	game.RenderSequence(
		func() render.Renderer {
			return render.NewEntitiesRenderer(atlas, camera, goke.Exclude[collisions.Hit]())
		},
		func() render.Renderer {
			return render.NewTagOverlayRenderer(atlas, camera, 0, color.RGBA{R: 255, A: 255}, goke.Include[collisions.Hit]())
		},
		func() render.Renderer {
			kin := &resources.Telemetry().Kinematics
			entityCount := func() int { return kin.DynamicCount + kin.StaticCount }
			return render.NewTelemetryRenderer(resources.TPS(), entityCount, &resources.Telemetry().Collision.Counter)
		},
	)

	game.Populate(EntityCount, &resources.Telemetry().Kinematics,
		kinematics.NewSpawner(
			grid.NewGridPlacement(ScreenWidth, ScreenHeight, RectSize),
			randomvelocity.New(200, 50, 10),
		),
		render.NewAppearanceExtras(func(index int) render.Appearance {
			return render.Appearance{
				Color:    color.RGBA{R: uint8(rand.IntN(206) + 50), G: uint8(rand.IntN(206) + 50), B: uint8(rand.IntN(206) + 50), A: 255},
				SpriteID: uint8(rand.IntN(procedural.SpriteCount)),
			}
		}),
	)

	game.Run()
}
