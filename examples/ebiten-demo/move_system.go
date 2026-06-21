package main

import (
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/spatial"
)

type MovementSystem struct {
	*Resources
	moveView *goke.View
	pos      goke.Col[Position]
	vel      goke.Col[Velocity]
}

var _ goke.System = (*MovementSystem)(nil)

func NewMoveSystem(resources *Resources) goke.System {
	return &MovementSystem{
		Resources: resources,
	}
}

func (s *MovementSystem) Init(ecs *goke.ECS) {
	s.moveView = goke.CreateView(ecs, goke.Track(&s.pos), goke.Track(&s.vel))
}

func (s *MovementSystem) Update(_ goke.Lookup, _ *goke.CmdBuf, d time.Duration) {
	dt := d.Seconds()
	s.moveView.All()
	for s.moveView.Next() {
		pos := s.pos.Slice(&s.moveView.Cursor)
		vel := s.vel.Slice(&s.moveView.Cursor)
		for i, entityID := range s.moveView.Cursor.IDs {
			pos[i].accX += float64(vel[i].X) * dt
			pos[i].accY += float64(vel[i].Y) * dt

			dx := int32(pos[i].accX)
			dy := int32(pos[i].accY)

			if dx != 0 {
				pos[i].accX -= float64(dx)
			}
			if dy != 0 {
				pos[i].accY -= float64(dy)
			}

			if dx != 0 || dy != 0 {
				delta := geom.NewVec(uint32(dx), uint32(dy))
				s.space.Translate(uint64(entityID), &pos[i].AABB, delta)
			}
		}
	}
	s.space.Flush(func(a spatial.AABB) {})
}
