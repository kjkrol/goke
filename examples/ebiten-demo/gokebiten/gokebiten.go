package gokebiten

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
)

const (
	defaultTargetTPS   = 60
	defaultPhysicsStep = time.Second / defaultTargetTPS
)

type GameStats struct {
	accumulator   time.Duration
	lastUpdate    time.Time
	Ticks         int       // Raw counter for physics steps
	lastTPSUpdate time.Time // Timer for resetting the counter
}

type GameProps struct {
	Title                     string
	TargetTPS                 int
	PhysicsStep               time.Duration
	ScreenWidth, ScreenHeight int
}

type Resources interface {
	GetGameProps() *GameProps
	Refresh(GameStats)
}

type Game[T Resources] struct {
	GameStats
	resources T
	ecs       *goke.ECS
	renderSeq []RenderSystem
}

var _ ebiten.Game = (*Game[Resources])(nil)

func NewGame[T Resources](resources T) *Game[T] {
	game := &Game[T]{
		resources: resources,
		GameStats: GameStats{},
		ecs:       goke.New(),
	}
	return game
}

func RegisterComponent[T Resources, C any](game *Game[T]) goke.ComponentDesc {
	return goke.RegisterComponent[C](game.ecs)
}

func (g *Game[T]) RegisterScheduledSystem(factory func(T) goke.System) goke.System {
	system := factory(g.resources)
	goke.RegisterSystem(g.ecs, system)
	return system
}

func (g *Game[T]) registerRenderSystem(factory func(T) RenderSystem) RenderSystem {
	system := factory(g.resources)
	system.Init(g.ecs)
	return system
}

func (g *Game[T]) RegisterSystem(factory func(T) goke.System) goke.System {
	system := factory(g.resources)
	system.Init(g.ecs)
	return system
}

func (g *Game[T]) LogicPlan(plan func(ctx goke.ExecutionContext, d time.Duration)) {
	goke.Plan(g.ecs, plan)
}

func (g *Game[T]) RenderSequence(sysFactories ...func(T) RenderSystem) {
	for _, factory := range sysFactories {
		renderSystem := g.registerRenderSystem(factory)
		g.renderSeq = append(g.renderSeq, renderSystem)
	}
}

func (g *Game[T]) Update() error {
	// 1. Calculate the real time elapsed since the last update (Delta Time)
	now := time.Now()
	if g.lastUpdate.IsZero() {
		g.lastUpdate = now
		g.lastTPSUpdate = now
	}
	elapsed := time.Since(g.lastUpdate)
	g.lastUpdate = now

	// 3. Add the elapsed time to the accumulator ("time bank")
	g.accumulator += elapsed

	// 4. Consume the accumulated time in fixed increments
	// If the frame rate drops, this loop will "catch up" by running multiple ticks
	maxSteps := 5 // Nie symuluj więcej niż 5 kroków na klatkę, nawet jak laguje
	steps := 0
	PhysicsStep := g.resources.GetGameProps().PhysicsStep
	for g.accumulator >= PhysicsStep && steps < maxSteps {
		goke.Tick(g.ecs, PhysicsStep)
		g.accumulator -= PhysicsStep
		g.Ticks++
		steps++
	}
	// Jeśli po 5 krokach nadal mamy "dług", po prostu go odpuszczamy,
	// żeby gra zwolniła (slow-motion), a nie klatkowała.
	if g.accumulator > PhysicsStep {
		g.accumulator = 0
	}

	// --- CALCULATE ACTUAL PHYSICS TPS ONCE PER SECOND ---
	if time.Since(g.lastTPSUpdate) >= time.Second {
		g.resources.Refresh(g.GameStats)
		g.Ticks = 0
		g.lastTPSUpdate = time.Now()
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
