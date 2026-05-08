package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

type StatsRendererSystem struct {
	*Resource
}

var _ gokebiten.RenderSystem = (*StatsRendererSystem)(nil)

func NewStatsRendererSystem(resource *Resource) *StatsRendererSystem {
	return &StatsRendererSystem{
		Resource: resource,
	}
}

func (s *StatsRendererSystem) Init(ecs *goke.ECS) {}

func (s *StatsRendererSystem) Update(_ goke.Lookup, _ *goke.Schedule, d time.Duration) {}

func (s *StatsRendererSystem) Draw(screen *ebiten.Image) {
	avgCollisionsPerTick := float64(0)
	if s.tps > 0 {
		avgCollisionsPerTick = float64(s.collisionCounter) / float64(s.tps)
	}
	debugMsg := fmt.Sprintf(
		"FPS: %0.2f\nTPS (Ebiten): %0.2f\nTPS (Physics): %d\nEntities: %d\nCollisions/Tick: %0.2f",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		s.tps,
		s.entityCount,
		avgCollisionsPerTick,
	)
	ebitenutil.DebugPrint(screen, debugMsg)
}
