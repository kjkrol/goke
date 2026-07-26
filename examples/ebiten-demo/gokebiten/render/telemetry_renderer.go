package render

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kjkrol/goke/v2"
)

var _ Renderer = (*TelemetryRenderer)(nil)

type TelemetryRenderer struct {
	measuredTPS              *int
	entityCount              func() int
	measuredCollisionCounter *int
}

func NewTelemetryRenderer(measuredTPS *int, entityCount func() int, measuredCollisionCounter *int) *TelemetryRenderer {
	return &TelemetryRenderer{
		measuredTPS:              measuredTPS,
		entityCount:              entityCount,
		measuredCollisionCounter: measuredCollisionCounter,
	}
}

func (s *TelemetryRenderer) Init(ecs *goke.ECS) {}

func (s *TelemetryRenderer) Draw(screen *ebiten.Image) {
	avgCollisionsPerTick := float64(0)
	if *s.measuredTPS > 0 {
		avgCollisionsPerTick = float64(*s.measuredCollisionCounter) / float64(*s.measuredTPS)
	}
	debugMsg := fmt.Sprintf(
		"FPS: %0.2f\nTPS (Ebiten): %0.2f\nTPS (Physics): %d\nEntities: %d\nCollision/Sec: %d\nCollisions/Tick: %0.2f",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		*s.measuredTPS,
		s.entityCount(),
		*s.measuredCollisionCounter,
		avgCollisionsPerTick,
	)
	ebitenutil.DebugPrint(screen, debugMsg)
}
