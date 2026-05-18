package main

import (
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
	"github.com/kjkrol/gokg/spatial"
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
}

func NewCollisionSystem(resouces *Resources) goke.System {
	return &CollisionSystem{
		Resources: resouces,
	}
}

func (s *CollisionSystem) Init(ecs *goke.ECS) {
	s.collisionView = goke.NewView3[Position, Velocity, Collision](ecs)
	s.compDescs.posDesc = goke.RegisterComponent[Position](ecs)
	s.compDescs.velDesc = goke.RegisterComponent[Velocity](ecs)
	s.compDescs.colDesc = goke.RegisterComponent[Collision](ecs)
}

func (s *CollisionSystem) Update(lookup goke.Lookup, sched *goke.Schedule, d time.Duration) {
	const solverIterations = 3
	const probeExpandMaring = 16
	var contacts []Contact = s.broadPhase(lookup, probeExpandMaring)
	s.narrowPhase(contacts, solverIterations)
	s.space.Flush(func(spatial.AABB) {})
}

func (s *CollisionSystem) broadPhase(lookup goke.Lookup, probeExpandMargin uint32) (contacts []Contact) {
	for head := range s.collisionView.All() {
		pos, vel, col := head.V1, head.V2, head.V3
		entityA := head.Entity

		if vel.X == 0 && vel.Y == 0 {
			continue
		}

		checkFunc := func(boxA geom.AABB[uint32], fragA plane.FragPosition) {
			s.space.Query(boxA, func(idB uint64, fragB plane.FragPosition) {
				entityB := goke.Entity(idB)

				if entityA.Index() < entityB.Index() {

					posBptr, _ := lookup.ComponentGet(entityB, s.compDescs.posDesc.ID)
					velBptr, _ := lookup.ComponentGet(entityB, s.compDescs.velDesc.ID)
					colBptr, _ := lookup.ComponentGet(entityB, s.compDescs.colDesc.ID)

					contacts = append(contacts, Contact{
						EntityA: entityA, EntityB: entityB,
						PosA: pos, PosB: (*Position)(posBptr),
						VelA: vel, VelB: (*Velocity)(velBptr),
						ColA: col, ColB: (*Collision)(colBptr),
						FragA: fragA, FragB: fragB,
					})
				}
			})
		}

		probeBoxA := pos.AABB
		s.space.ExpandOnly(&probeBoxA, probeExpandMargin)
		checkFunc(probeBoxA.AABB, plane.FRAG_MAIN)
		probeBoxA.VisitFragments(func(fragA plane.FragPosition, boxA geom.AABB[uint32]) bool {
			checkFunc(boxA, fragA)
			return true
		})
	}
	return
}

func (s *CollisionSystem) narrowPhase(contacts []Contact, solverIterations int) {
	now := time.Now()

	for iter := 0; iter < solverIterations; iter++ {
		for _, c := range contacts {

			boxA := c.freshBoxA()
			boxB := c.freshBoxB()
			penetrationVec := c.penetration(boxA, boxB)

			if penetrationVec.X == 0 && penetrationVec.Y == 0 {
				continue
			}

			if !c.resolved {
				c.resolveVelocity(penetrationVec)
				c.updateStats(now)
				s.collisionCounter++
				c.resolved = true
			}

			if mtv1, mtv2, ok := c.calculateMtv(boxA, boxB, false); ok == true {
				s.space.Translate(uint64(c.EntityA.Index()), &c.PosA.AABB, mtv1)
				s.space.Translate(uint64(c.EntityB.Index()), &c.PosB.AABB, mtv2)
			}
		}
	}
}

// ----- CONTACT -----

type Contact struct {
	EntityA goke.Entity
	EntityB goke.Entity
	PosA    *Position
	PosB    *Position
	VelA    *Velocity
	VelB    *Velocity
	ColA    *Collision
	ColB    *Collision

	FragA    plane.FragPosition
	FragB    plane.FragPosition
	resolved bool
}

func (c *Contact) updateStats(now time.Time) {
	c.ColA.timestamp = now
	c.ColA.counter++
	c.ColB.timestamp = now
	c.ColB.counter++
}

func (c *Contact) freshBoxA() geom.AABB[uint32] {
	if c.FragA == plane.FRAG_MAIN {
		return c.PosA.AABB.AABB
	}

	var freshBox geom.AABB[uint32]
	c.PosA.AABB.VisitFragments(func(fp plane.FragPosition, b geom.AABB[uint32]) bool {
		if fp == c.FragA {
			freshBox = b
			return false
		}
		return true
	})
	return freshBox
}

func (c *Contact) freshBoxB() geom.AABB[uint32] {
	if c.FragB == plane.FRAG_MAIN {
		return c.PosB.AABB.AABB
	}

	var freshBox geom.AABB[uint32]
	c.PosB.AABB.VisitFragments(func(fp plane.FragPosition, b geom.AABB[uint32]) bool {
		if fp == c.FragB {
			freshBox = b
			return false
		}
		return true
	})
	return freshBox
}

// resolveVelocity przetwarza fizykę odbicia AABB
func (c *Contact) resolveVelocity(pen geom.Vec[int32]) {
	if pen.X != 0 {
		relVelX := int32(c.VelA.Vec.X) - int32(c.VelB.Vec.X)

		if (pen.X > 0 && relVelX > 0) || (pen.X < 0 && relVelX < 0) {
			return
		}

		tempX := c.VelA.Vec.X
		c.VelA.Vec.X = c.VelB.Vec.X
		c.VelB.Vec.X = tempX
	} else if pen.Y != 0 {
		relVelY := int32(c.VelA.Vec.Y) - int32(c.VelB.Vec.Y)

		if (pen.Y > 0 && relVelY > 0) || (pen.Y < 0 && relVelY < 0) {
			return
		}

		tempY := c.VelA.Vec.Y
		c.VelA.Vec.Y = c.VelB.Vec.Y
		c.VelB.Vec.Y = tempY
	}
}

// calculate minimum translation vector
func (c *Contact) calculateMtv(r1, r2 geom.AABB[uint32], isStaticB bool) (mtv1, mtv2 geom.Vec[uint32], res bool) {
	pen := c.penetration(r1, r2)

	if pen.X == 0 && pen.Y == 0 {
		return geom.Vec[uint32]{}, geom.Vec[uint32]{}, false
	}

	calculatePush := func(penetration int32) (int32, int32) {
		if isStaticB {
			return penetration, 0
		} else {
			pA := penetration / 2
			pB := -(penetration - pA) // Zapewnia brak zgubionego piksela przy liczbach nieparzystych
			return pA, pB
		}
	}

	var pushA, pushB geom.Vec[int32]

	if pen.X != 0 {
		p1, p2 := calculatePush(pen.X)
		pushA = geom.Vec[int32]{X: p1, Y: 0}
		pushB = geom.Vec[int32]{X: p2, Y: 0}
	} else {
		p1, p2 := calculatePush(pen.Y)
		pushA = geom.Vec[int32]{X: 0, Y: p1}
		pushB = geom.Vec[int32]{X: 0, Y: p2}
	}

	mtv1 = geom.NewVec(uint32(pushA.X), uint32(pushA.Y))
	mtv2 = geom.NewVec(uint32(pushB.X), uint32(pushB.Y))
	res = true
	return
}

func (C *Contact) penetration(r1, r2 geom.AABB[uint32]) geom.Vec[int32] {
	// 1. Sprawdzamy czy w ogóle jest kolizja (czyste rzuty osi)
	leftX := max(int32(r1.TopLeft.X), int32(r2.TopLeft.X))
	rightX := min(int32(r1.BottomRight.X), int32(r2.BottomRight.X))
	overlapX := rightX - leftX
	if overlapX <= 0 {
		return geom.Vec[int32]{}
	}

	topY := max(int32(r1.TopLeft.Y), int32(r2.TopLeft.Y))
	bottomY := min(int32(r1.BottomRight.Y), int32(r2.BottomRight.Y))
	overlapY := bottomY - topY
	if overlapY <= 0 {
		return geom.Vec[int32]{}
	}

	// 2. Obliczamy dystans ucieczki w 4 kierunkach (jak daleko r1 musi się przesunąć, by minąć r2)
	pushRight := int32(r2.BottomRight.X) - int32(r1.TopLeft.X)
	pushLeft := int32(r1.BottomRight.X) - int32(r2.TopLeft.X)
	pushDown := int32(r2.BottomRight.Y) - int32(r1.TopLeft.Y)
	pushUp := int32(r1.BottomRight.Y) - int32(r2.TopLeft.Y)

	// 3. Wybieramy najkrótszą możliwą drogę ucieczki
	minPush := pushRight
	mtv := geom.Vec[int32]{X: pushRight, Y: 0}

	if pushLeft < minPush {
		minPush = pushLeft
		mtv = geom.Vec[int32]{X: -pushLeft, Y: 0}
	}
	if pushDown < minPush {
		minPush = pushDown
		mtv = geom.Vec[int32]{X: 0, Y: pushDown}
	}
	if pushUp < minPush {
		mtv = geom.Vec[int32]{X: 0, Y: -pushUp}
	}

	return mtv
}
