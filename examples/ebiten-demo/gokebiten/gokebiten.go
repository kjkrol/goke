package gokebiten

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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

type Resources interface {
	GetGameProps() *GameProps
	GetInputEvents() *InputEvents
	Refresh(int)
	Reset()
}

type Game[T Resources] struct {
	ticks        int
	physicsStep  time.Duration
	timeTracker  *TimeTracker
	resources    T
	ecs          *goke.ECS
	renderSeq    []RenderSystem
	inputAdapter InputAdapter
}

var _ ebiten.Game = (*Game[Resources])(nil)

func NewGame[T Resources](resources T, inputAdapter InputAdapter) *Game[T] {
	if inputAdapter == nil {
		panic("InputAdapter has to be set")
	}
	targetTPS := defaultTargetTPS
	if resources.GetGameProps() != nil && resources.GetGameProps().TargetTPS != 0 {
		targetTPS = resources.GetGameProps().TargetTPS
	}
	game := &Game[T]{
		resources:    resources,
		physicsStep:  time.Second / time.Duration(targetTPS),
		timeTracker:  NewTimeTracker(),
		ecs:          goke.New(),
		inputAdapter: inputAdapter,
	}
	return game
}

func RegComp[T Resources, C any](game *Game[T]) goke.CompID {
	return goke.RegComp[C](game.ecs)
}

func (g *Game[T]) RegisterScheduledSystem(factory func(T) goke.System) goke.System {
	system := factory(g.resources)
	g.ecs.RegSys(system)
	return system
}

func (g *Game[T]) registerRenderSystem(factory func(T) RenderSystem) RenderSystem {
	system := factory(g.resources)
	system.Init(g.ecs)
	return system
}

func (g *Game[T]) RegSys(factory func(T) goke.System) goke.System {
	system := factory(g.resources)
	system.Init(g.ecs)
	return system
}

func (g *Game[T]) LogicPlan(plan func(ctx goke.RunCtx, d time.Duration)) {
	g.ecs.SetPlan(plan)
}

func (g *Game[T]) RenderSequence(sysFactories ...func(T) RenderSystem) {
	for _, factory := range sysFactories {
		renderSystem := g.registerRenderSystem(factory)
		g.renderSeq = append(g.renderSeq, renderSystem)
	}
}

func (g *Game[T]) Update() error {
	g.inputAdapter.Capture(g.resources.GetInputEvents())

	steps := g.timeTracker.CalculateSteps(g.physicsStep, 5)
	for range steps {
		g.ecs.Tick(g.physicsStep)
		g.resources.GetInputEvents().ResetTransient()
		g.ticks++
	}

	if g.timeTracker.ProcessStatsInterval() {
		g.resources.Refresh(g.ticks)
		g.ticks = 0
		g.resources.Reset()
	}

	return nil
}

func (g *Game[T]) Draw(screen *ebiten.Image) {
	for _, sys := range g.renderSeq {
		sys.Draw(screen)
	}
}

func (g *Game[T]) Layout(outsideWidth, outsideHeight int) (int, int) {
	ScreenWidth := g.resources.GetGameProps().ScreenWidth
	ScreenHeight := g.resources.GetGameProps().ScreenHeight
	return ScreenWidth, ScreenHeight
}

func (g *Game[T]) Run() {
	ScreenWidth := g.resources.GetGameProps().ScreenWidth
	ScreenHeight := g.resources.GetGameProps().ScreenHeight
	Title := g.resources.GetGameProps().Title
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle(Title)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
