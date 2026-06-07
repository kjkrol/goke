package main

import (
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/spatial"
)

type MovementSystem struct {
	*Resources
	moveView *goke.View2[Position, Velocity]
}

var _ goke.System = (*MovementSystem)(nil)

func NewMoveSystem(resources *Resources) goke.System {
	return &MovementSystem{
		Resources: resources,
	}
}

func (s *MovementSystem) Init(ecs *goke.ECS) {
	s.moveView = goke.NewView2[Position, Velocity](ecs)
}

func (s *MovementSystem) Update(_ goke.Lookup, _ *goke.Schedule, d time.Duration) {
	dt := d.Seconds()
	for page := range s.moveView.All() {
		for i, entity := range page.Entity {
			pos, vel := &page.Comp1[i], &page.Comp2[i]
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
				s.space.Translate(uint64(entity), &pos.AABB, delta)
			}
		}
	}
	s.space.Flush(func(a spatial.AABB) {})
}
