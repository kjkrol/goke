package main

import (
	"time"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

// --- Configuration ---
const (
	TPS            = 60
	ScreenWidth    = 1024
	ScreenHeight   = 1024
	RectSize       = 4 // TODO: to powinno byc pobierane z AABB encji
	BucketCapacity = 32
	EntityCount    = 64 * 1 // to nie jest zaden count, ale init liczba encji
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
	grid := NewGrid(ScreenWidth, ScreenHeight, BucketCapacity, EntityCount)
	gameProps := gokebiten.GameProps{
		Title:        "GOKe + GOKg + Ebiten Integration",
		ScreenWidth:  ScreenWidth,
		ScreenHeight: ScreenHeight,
		TargetTPS:    TPS,
		PhysicsStep:  time.Second / time.Duration(TPS),
	}
	return &Resource{
		gameProps:   gameProps,
		grid:        grid,
		rectSize:    RectSize,
		entityCount: EntityCount,
	}
}
