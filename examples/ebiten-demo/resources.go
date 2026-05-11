package main

import (
	"time"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/gokg/pkg/spatial"
)

// --- Configuration ---
const (
	TPS            = 60
	ScreenWidth    = 1024
	ScreenHeight   = 1024
	RectSize       = 10
	BucketCapacity = 32
	EntityCount    = 256
)

type Statistics struct {
	collisionCounter int
	tps              int
	entityCount      int
}

type Resources struct {
	gameProps *gokebiten.GameProps
	grid      *Grid
	rectSize  int
	Statistics
}

var _ (gokebiten.Resources) = (*Resources)(nil)

func NewResources() *Resources {
	return &Resources{
		gameProps: &gokebiten.GameProps{
			Title:        "GOKe + GOKg + Ebiten Integration",
			ScreenWidth:  ScreenWidth,
			ScreenHeight: ScreenHeight,
			TargetTPS:    TPS,
			PhysicsStep:  time.Second / time.Duration(TPS),
		},
		grid: NewGrid(spatial.GridIndexConfig{
			Resolution:       spatial.Size1024x1024,
			BucketResolution: spatial.Size64x64,
			BucketCapacity:   BucketCapacity,
			OpsBufferSize:    EntityCount,
		}),
		rectSize: RectSize,
		Statistics: Statistics{
			entityCount: EntityCount,
		},
	}
}

func (r *Resources) GetGameProps() *gokebiten.GameProps {
	return r.gameProps
}

func (r *Resources) Refresh(gs gokebiten.GameStats) {
	r.collisionCounter = 0
	r.tps = gs.Ticks
}
