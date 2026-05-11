package main

import (
	"math"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/plane"
	"github.com/kjkrol/gokg/pkg/spatial"
)

var _ goke.System = (*CollisionSystem)(nil)

type CollisionSystem struct {
	*Resources
	collisionView *goke.View3[Position, Velocity, Collision]
	compDescs     struct {
		posDesc goke.ComponentDesc
		velDesc goke.ComponentDesc
		appDesc goke.ComponentDesc
		colDesc goke.ComponentDesc
	}
	compensator penetrationCompensator
}

func NewCollisionSystem(resouces *Resources) goke.System {
	return &CollisionSystem{
		Resources:   resouces,
		compensator: penetrationCompensator{resouces.grid.space},
	}
}

func (s *CollisionSystem) Init(ecs *goke.ECS) {
	s.collisionView = goke.NewView3[Position, Velocity, Collision](ecs)
	s.compDescs.posDesc = goke.RegisterComponent[Position](ecs)
	s.compDescs.velDesc = goke.RegisterComponent[Velocity](ecs)
	s.compDescs.colDesc = goke.RegisterComponent[Collision](ecs)
}

func (s *CollisionSystem) Update(lookup goke.Lookup, sched *goke.Schedule, d time.Duration) {
	for head := range s.collisionView.All() {
		pos, vel, col := head.V1, head.V2, head.V3

		if vel.X == 0 && vel.Y == 0 {
			return
		}

		now := time.Now()
		s.grid.spatialIndex.QueryRange(pos.AABB.AABB, func(otherID uint64) {
			entity2 := goke.Entity(otherID)
			if head.Entity.Index() < entity2.Index() {

				pos2ptr, _ := lookup.ComponentGet(entity2, s.compDescs.posDesc.ID)
				vel2ptr, _ := lookup.ComponentGet(entity2, s.compDescs.velDesc.ID)
				col2ptr, _ := lookup.ComponentGet(entity2, s.compDescs.colDesc.ID)

				pos2 := (*Position)(pos2ptr)
				vel2 := (*Velocity)(vel2ptr)
				col2 := (*Collision)(col2ptr)

				col.timestamp = now
				col.counter++
				col2.timestamp = now
				col2.counter++

				s.collisionCounter++

				// if col.counter > 3 {
				// 	goke.ScheduleRemoveEntity(sched, head.Entity)
				// 	s.grid.spatialIndex.QueueRemove(uint64(head.Entity))
				// }
				// if col2.counter > 3 {
				// 	goke.ScheduleRemoveEntity(sched, entity2)
				// 	s.grid.spatialIndex.QueueRemove(uint64(entity2))
				// }
				// if col.counter > 3 || col2.counter > 3 {
				// 	return
				// }

				s.resolveCollision(pos, vel, pos2, vel2)

				s.grid.spatialIndex.QueueUpdate(uint64(head.Entity), pos.AABB.AABB, true)
				s.grid.spatialIndex.QueueUpdate(otherID, pos2.AABB.AABB, true)
			}
		})
	}
	s.grid.spatialIndex.Flush(func(a spatial.AABB) {})
}

func (s *CollisionSystem) resolveCollision(
	pos1 *Position, vel1 *Velocity,
	pos2 *Position, vel2 *Velocity,
) {
	// swap velocity
	tempVel := vel1.Vec
	vel1.Vec = vel2.Vec
	vel2.Vec = tempVel

	s.compensator.compensate(pos1, pos2)
}

type penetrationCompensator struct {
	space plane.Space2D[uint32]
}

func (p penetrationCompensator) compensate(pos1, pos2 *Position) {
	if mtv1, mtv2, ok := p.minimumTranslationVector(pos1, pos2); ok == true {
		p.space.Translate(&pos1.AABB, mtv1)
		p.space.Translate(&pos2.AABB, mtv2)
	}
}

func (p penetrationCompensator) minimumTranslationVector(pos1, pos2 *Position) (mtv1, mtv2 geom.Vec[uint32], res bool) {
	pen := p.penetration(pos1, pos2)

	if pen.X == 0 || pen.Y == 0 {
		res = false
		return
	}
	if math.Abs(float64(pen.X)) < math.Abs(float64(pen.Y)) {
		push := pen.X / 2
		if push == 0 {
			push = pen.X // Zabezpieczenie przed 0 przy penetracji o 1 piksel
		}
		mtv1 = geom.NewVec(uint32(-push), 0)
		mtv2 = geom.NewVec(uint32(push), 0)
	} else {
		push := pen.Y / 2
		if push == 0 {
			push = pen.Y // Zabezpieczenie
		}
		mtv1 = geom.NewVec(0, uint32(-push))
		mtv2 = geom.NewVec(0, uint32(push))
	}
	res = true
	return
}

func (p *penetrationCompensator) penetration(pos1, pos2 *Position) geom.Vec[int32] {
	r1 := pos1.AABB.AABB
	r2 := pos2.AABB.AABB

	var directX int32 = 1
	var directY int32 = 1

	var penX, penY int32
	if r1.TopLeft.X < r2.TopLeft.X {
		penX = int32(r1.BottomRight.X) - int32(r2.TopLeft.X)
	} else {
		penX = int32(r2.BottomRight.X) - int32(r1.TopLeft.X)
		directX = -1
	}

	if penX <= 0 {
		return geom.Vec[int32]{X: 0, Y: 0}
	}

	if r1.TopLeft.Y < r2.TopLeft.Y {
		penY = int32(r1.BottomRight.Y) - int32(r2.TopLeft.Y)
	} else {
		penY = int32(r2.BottomRight.Y) - int32(r1.TopLeft.Y)
		directY = -1
	}

	if penY <= 0 {
		return geom.Vec[int32]{X: 0, Y: 0}
	}

	return geom.Vec[int32]{
		X: penX * directX,
		Y: penY * directY,
	}
}
