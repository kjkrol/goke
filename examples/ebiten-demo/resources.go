package main

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/gokg"
	"github.com/kjkrol/gokg/spatial"
)

// --- Configuration ---
const (
	TPS            = 2 * 60
	ScreenWidth    = 1024
	ScreenHeight   = 1024
	RectSize       = 16
	BucketCapacity = 64
	EntityCount    = 256
)

type Statistics struct {
	collisionCounter          int
	entityCounter             int
	measuredTPS               int
	meeasuredCollisionCounter int
}

type Resources struct {
	gameProps *gokebiten.GameProps
	space     *gokg.Space
	rectSize  int
	inputs    gokebiten.InputEvents
	Statistics
}

var _ (gokebiten.Resources) = (*Resources)(nil)

func NewResources() *Resources {
	space, _ := gokg.NewSpace(gokg.Config{
		Width:          ScreenWidth,
		Height:         ScreenHeight,
		Toroidal:       true,
		BucketSize:     spatial.Size64x64,
		BucketCapacity: BucketCapacity,
		OpsBufferSize:  EntityCount * 4,
	})
	return &Resources{
		gameProps: &gokebiten.GameProps{
			Title:        "GOKe + GOKg + Ebiten Integration",
			ScreenWidth:  ScreenWidth,
			ScreenHeight: ScreenHeight,
			TargetTPS:    TPS,
		},
		space:    space,
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
