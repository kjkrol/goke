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
	EntityCount    = 128
)

type Statistics struct {
	collisionCounter int
	tps              int
}

type Resource struct {
	gameProps   gokebiten.GameProps
	grid        *Grid
	rectSize    int
	entityCount int
	Statistics
}

func NewResource() *Resource {
	gridConfig := spatial.GridIndexConfig{
		Resolution:       spatial.Size1024x1024,
		BucketResolution: spatial.Size64x64,
		BucketCapacity:   BucketCapacity,
		OpsBufferSize:    EntityCount,
	}
	gameProps := gokebiten.GameProps{
		Title:        "GOKe + GOKg + Ebiten Integration",
		ScreenWidth:  ScreenWidth,
		ScreenHeight: ScreenHeight,
		TargetTPS:    TPS,
		PhysicsStep:  time.Second / time.Duration(TPS),
	}
	return &Resource{
		gameProps:   gameProps,
		grid:        NewGrid(gridConfig),
		rectSize:    RectSize,
		entityCount: EntityCount,
	}
}
