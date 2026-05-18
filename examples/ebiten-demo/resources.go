package main

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/gokg/pkg/spatial"
)

// --- Configuration ---
const (
	TPS            = 60
	ScreenWidth    = 1024
	ScreenHeight   = 1024
	RectSize       = 90
	BucketCapacity = 16
	EntityCount    = 32
)

type Statistics struct {
	collisionCounter          int
	entityCounter             int
	measuredTPS               int
	meeasuredCollisionCounter int
}

type Resources struct {
	gameProps *gokebiten.GameProps
	grid      *Grid
	rectSize  int
	inputs    gokebiten.InputEvents
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
		},
		grid: NewGrid(spatial.GridIndexConfig{
			Resolution:       spatial.Size1024x1024,
			BucketResolution: spatial.Size64x64,
			BucketCapacity:   BucketCapacity,
			OpsBufferSize:    EntityCount,
		}),
		rectSize: RectSize,
		Statistics: Statistics{
			entityCounter: EntityCount,
		},
	}
}

func (r *Resources) GetGameProps() *gokebiten.GameProps {
	return r.gameProps
}

func (r *Resources) Reset() {
	r.collisionCounter = 0
}

func (r *Resources) Refresh(tick int) {
	r.measuredTPS = tick
	r.meeasuredCollisionCounter = r.collisionCounter
}

func (r *Resources) GetInputEvents() *gokebiten.InputEvents { return &r.inputs }
