package main

import (
	"image/color"
	"math"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/spatial"
)

var _ goke.System = (*CollisionSystem)(nil)

type CollisionSystem struct {
	*Resource
	collisionView *goke.View3[Position, Velocity, Appearance]
}

func NewCollisionSystem(resouce *Resource) *CollisionSystem {
	return &CollisionSystem{
		Resource: resouce,
	}
}

func (s *CollisionSystem) Init(ecs *goke.ECS) {
	s.collisionView = goke.NewView3[Position, Velocity, Appearance](ecs)
}

func (s *CollisionSystem) Update(lookup goke.Lookup, _ *goke.Schedule, d time.Duration) {
	for head := range s.collisionView.All() {
		pos, vel, app := head.V1, head.V2, head.V3

		s.grid.spatialIndex.QueryRange(pos.AABB.AABB, func(otherID uint64) {
			entity2 := goke.Entity(otherID)
			if head.Entity.Index() < entity2.Index() {

				pos2, _ := lookup.ComponentGet(entity2, posDesc.ID)
				vel2, _ := lookup.ComponentGet(entity2, velDesc.ID)
				app2, _ := lookup.ComponentGet(entity2, appDesc.ID)

				app.Color = color.RGBA{R: 255, A: 255}
				(*Appearance)(app2).Color = color.RGBA{R: 255, A: 255}
				s.resolveCollision(pos, vel, (*Position)(pos2), (*Velocity)(vel2))
				s.collisionCounter++

				s.grid.spatialIndex.QueueUpdate(uint64(head.Entity), pos.AABB.AABB, true)
				s.grid.spatialIndex.QueueUpdate(otherID, (*Position)(pos2).AABB.AABB, true)
			}
		})
	}
	s.grid.spatialIndex.Flush(func(a spatial.AABB) {})
}

func (s *CollisionSystem) resolveCollision(
	pos1 *Position, vel1 *Velocity,
	pos2 *Position, vel2 *Velocity,
) bool {
	pen := s.penetration(pos1, pos2)

	if pen.X == 0 || pen.Y == 0 {
		return false
	}

	// swap velocity
	tempVel := vel1.Vec
	vel1.Vec = vel2.Vec
	vel2.Vec = tempVel

	var mtv1, mtv2 geom.Vec[uint32]

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

	s.grid.space.Translate(&pos1.AABB, mtv1)
	s.grid.space.Translate(&pos2.AABB, mtv2)

	return true
}

func (s *CollisionSystem) penetration(pos1, pos2 *Position) geom.Vec[int32] {
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
