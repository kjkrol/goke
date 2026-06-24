package main

import (
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/spatial"
)

type MovementSystem struct {
	*Resources
	moveQuery *goke.Query
	pos       goke.Comp[Position]
	vel       goke.Comp[Velocity]
}

var _ goke.System = (*MovementSystem)(nil)

func NewMoveSystem(resources *Resources) goke.System {
	return &MovementSystem{
		Resources: resources,
	}
}

func (s *MovementSystem) Init(ecs *goke.ECS) {
	s.moveQuery = ecs.NewQueryBuilder(&s.pos, &s.vel).Build()
}

func (s *MovementSystem) Update(_ *goke.CmdBuf, d time.Duration) {
	dt := d.Seconds()
	s.moveQuery.All()
	for s.moveQuery.Next() {
		pos := s.pos.Slice(&s.moveQuery.Cursor)
		vel := s.vel.Slice(&s.moveQuery.Cursor)
		for i, entityID := range s.moveQuery.Cursor.IDs {
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
