package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/goke/v2"
)

type StatsRendererSystem struct {
	*Resources
}

var _ gokebiten.RenderSystem = (*StatsRendererSystem)(nil)

func NewStatsRendererSystem(resources *Resources) gokebiten.RenderSystem {
	return &StatsRendererSystem{
		Resources: resources,
	}
}

func (s *StatsRendererSystem) Init(ecs *goke.ECS) {}

func (s *StatsRendererSystem) Update(_ *goke.CmdBuf, d time.Duration) {}

func (s *StatsRendererSystem) Draw(screen *ebiten.Image) {
	avgCollisionsPerTick := float64(0)
	if s.measuredTPS > 0 {
		avgCollisionsPerTick = float64(s.meeasuredCollisionCounter) / float64(s.measuredTPS)
	}
	debugMsg := fmt.Sprintf(
		"FPS: %0.2f\nTPS (Ebiten): %0.2f\nTPS (Physics): %d\nEntities: %d\nCollision/Sec: %d\nCollisions/Tick: %0.2f",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		s.measuredTPS,
		s.entityCounter,
		s.meeasuredCollisionCounter,
		avgCollisionsPerTick,
	)
	ebitenutil.DebugPrint(screen, debugMsg)
}
