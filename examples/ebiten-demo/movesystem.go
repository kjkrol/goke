package main

import (
	"image/color"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/spatial"
)

type MovementSystem struct {
	*Resource
	moveView *goke.View3[Position, Velocity, Appearance]
}

var _ goke.System = (*MovementSystem)(nil)

func NewMoveSystem(resource *Resource) *MovementSystem {
	return &MovementSystem{
		Resource: resource,
	}
}

func (s *MovementSystem) Init(ecs *goke.ECS) {
	s.moveView = goke.NewView3[Position, Velocity, Appearance](ecs)
}

func (s *MovementSystem) Update(_ goke.Lookup, _ *goke.Schedule, d time.Duration) {
	dt := d.Seconds()
	for head := range s.moveView.All() {
		pos, vel := head.V1, head.V2
		app := head.V3
		app.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		pos.accX += float64(vel.X) * dt
		pos.accY += float64(vel.Y) * dt

		dx := int32(pos.accX)
		dy := int32(pos.accY)

		if dx != 0 {
			pos.accX -= float64(dx)
		}
		if dy != 0 {
			pos.accY -= float64(dy)
		}

		if dx != 0 || dy != 0 {
			delta := geom.NewVec(uint32(dx), uint32(dy))
			s.grid.space.Translate(&pos.AABB, delta)
			s.grid.spatialIndex.QueueUpdate(uint64(head.Entity), pos.AABB.AABB, true)
		}
	}
	s.grid.spatialIndex.Flush(func(a spatial.AABB) {})
}
