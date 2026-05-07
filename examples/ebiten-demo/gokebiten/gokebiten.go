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

type Game struct {
	GameProps
	GameStats
	ecs           *goke.ECS
	renderPlan    func(screen *ebiten.Image)
	extraFnPerSec func(GameStats)
}

var _ ebiten.Game = (*Game)(nil)

type GameOption func(*Game)

func WithExtraFnPerSec(fn func(GameStats)) GameOption {
	return func(g *Game) { g.extraFnPerSec = fn }
}

func NewGame(props GameProps, opts ...GameOption) *Game {
	game := &Game{
		GameProps: props,
		GameStats: GameStats{},
		ecs:       goke.New(),
	}
	for _, option := range opts {
		option(game)
	}
	return game
}

func (g *Game) RegisterScheduledSystem(system goke.System) {
	goke.RegisterSystem(g.ecs, system)
}

func (g *Game) RegisterRenderSystem(system RenderSystem) {
	system.Init(g.ecs)
}

func (g *Game) RegisterSystem(system goke.System) {
	system.Init(g.ecs)
}

func (g *Game) LogicPlan(plan func(ctx goke.ExecutionContext, d time.Duration)) {
	goke.Plan(g.ecs, plan)
}

func (g *Game) RenderPlan(plan func(screen *ebiten.Image)) {
	g.renderPlan = plan
}

func (g *Game) Update() error {
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
	for g.accumulator >= g.PhysicsStep && steps < maxSteps {
		goke.Tick(g.ecs, g.PhysicsStep)
		g.accumulator -= g.PhysicsStep
		g.Ticks++
		steps++
	}
	// Jeśli po 5 krokach nadal mamy "dług", po prostu go odpuszczamy,
	// żeby gra zwolniła (slow-motion), a nie klatkowała.
	if g.accumulator > g.PhysicsStep {
		g.accumulator = 0
	}

	// --- CALCULATE ACTUAL PHYSICS TPS ONCE PER SECOND ---
	if time.Since(g.lastTPSUpdate) >= time.Second {
		g.extraFnPerSec(g.GameStats)
		g.Ticks = 0
		g.lastTPSUpdate = time.Now()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderPlan(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.ScreenWidth, g.ScreenHeight
}

func (g *Game) Run() {
	ebiten.SetWindowSize(g.ScreenWidth, g.ScreenHeight)
	ebiten.SetWindowTitle(g.Title)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
