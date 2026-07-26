package gokebiten

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/control"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/render"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/spatial"
	"github.com/kjkrol/goke/v2"
)

const (
	defaultTargetTPS = 60
)

type GameProps struct {
	Title                     string
	TargetTPS                 int
	ScreenWidth, ScreenHeight int
}

// resources is the narrow contract Game itself needs — satisfied by
// *Resources[S, T] for any S, T. Unexported: Resources is the only, and only
// intended, implementation.
type resources interface {
	GetGameProps() *GameProps
	GetInputEvents() *control.InputEvents
	GetSpaceConfig() spatial.Config
	TPS() *int
	Reset()
}

type Game struct {
	ticks             int
	physicsStep       time.Duration
	timeTracker       *TimeTracker
	resources         resources
	ecs               *goke.ECS
	renderSeq         []render.Renderer
	controller        *control.DefaultController
	spaceConfig       spatial.Config
	world             *spatial.World
	kinematicsEnabled bool
	paused            bool
}

var _ ebiten.Game = (*Game)(nil)

func NewGame(res resources) *Game {
	targetTPS := defaultTargetTPS
	if res.GetGameProps() != nil && res.GetGameProps().TargetTPS != 0 {
		targetTPS = res.GetGameProps().TargetTPS
	}
	controller := control.NewDefaultController(&control.DesktopAdapter{}, res.GetInputEvents())
	game := &Game{
		resources:   res,
		physicsStep: time.Second / time.Duration(targetTPS),
		timeTracker: NewTimeTracker(),
		ecs:         goke.New(),
		controller:  controller,
		spaceConfig: res.GetSpaceConfig(),
	}
	game.ecs.RegSys(controller)
	return game
}

func (g *Game) SetEventHandler(handler control.EventHandler) {
	g.controller.SetHandler(handler)
}

func (g *Game) Paused() bool { return g.paused }

func (g *Game) Pause() { g.paused = true }

func (g *Game) Resume() { g.paused = false }

func (g *Game) TogglePause() { g.paused = !g.paused }

func RegComp[C any](game *Game) goke.CompID {
	return goke.RegComp[C](game.ecs)
}

func (g *Game) RegSys(factory func() goke.System) goke.System {
	system := factory()
	g.ecs.RegSys(system)
	return system
}

func (g *Game) registerRenderer(factory func() render.Renderer) render.Renderer {
	renderer := factory()
	renderer.Init(g.ecs)
	return renderer
}

func (g *Game) Setup(factories ...func() spatial.Setup) {
	for _, factory := range factories {
		factory().Init(g.ecs)
	}
}

func (g *Game) Loop(plan func(ctx goke.RunCtx, d time.Duration)) {
	g.ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(g.controller, d)
		if g.paused {
			return
		}
		plan(ctx, d)
	})
}

func (g *Game) RenderSequence(rendererFactories ...func() render.Renderer) {
	for _, factory := range rendererFactories {
		renderer := g.registerRenderer(factory)
		g.renderSeq = append(g.renderSeq, renderer)
	}
}

// UseWorld builds the spatial index from this Game's SpaceConfig,
// provisioned for pop. Must be called before Populate, PopulateStatic or
// UsePhysics.
func (g *Game) UseWorld(pop spatial.Population) {
	g.world = spatial.NewWorld(g.ecs, g.spaceConfig, pop)
}

// Populate spawns count moving entities — see spatial.World.Populate.
// Requires UseWorld and UsePhysics.
func (g *Game) Populate(count int, telemetry *kinematics.Telemetry, populators ...spatial.EntityExtras) {
	if g.world == nil {
		panic("gokebiten: UseWorld must be called before Populate")
	}
	if !g.kinematicsEnabled {
		panic("gokebiten: UsePhysics must be called before Populate")
	}
	g.world.Populate(count, telemetry, populators...)
}

// PopulateStatic spawns count static entities — see spatial.World.PopulateStatic.
// Requires UseWorld.
func (g *Game) PopulateStatic(count int, telemetry *kinematics.Telemetry, populators ...spatial.EntityExtras) {
	if g.world == nil {
		panic("gokebiten: UseWorld must be called before PopulateStatic")
	}
	g.world.PopulateStatic(count, telemetry, populators...)
}

// UsePhysics builds the Physics wrapper (kinematics + collisions) for this
// world. Physics needs no reference to Game — it gets exactly what it needs
// as explicit parameters.
func (g *Game) UsePhysics() *physics.Physics {
	if g.world == nil {
		panic("gokebiten: UseWorld must be called before UsePhysics")
	}
	g.kinematicsEnabled = true
	return physics.New(g.world.Space(), g.ecs, g.world.Population().MinSize, g.physicsStep)
}

func (g *Game) Update() error {
	g.controller.Capture(g.resources.GetInputEvents())

	steps := g.timeTracker.CalculateSteps(g.physicsStep, 5)
	for range steps {
		g.ecs.Tick(g.physicsStep)
		g.resources.GetInputEvents().ResetTransient()
		g.ticks++
	}

	if g.timeTracker.ProcessStatsInterval() {
		*g.resources.TPS() = g.ticks
		g.ticks = 0
		g.resources.Reset()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, sys := range g.renderSeq {
		sys.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	ScreenWidth := g.resources.GetGameProps().ScreenWidth
	ScreenHeight := g.resources.GetGameProps().ScreenHeight
	return ScreenWidth, ScreenHeight
}

func (g *Game) Run() {
	ScreenWidth := g.resources.GetGameProps().ScreenWidth
	ScreenHeight := g.resources.GetGameProps().ScreenHeight
	Title := g.resources.GetGameProps().Title
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle(Title)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
