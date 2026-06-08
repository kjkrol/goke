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
	collisionView    *goke.View3[Position, Velocity, Collision]
	contactsBuffer   []Contact
	contactsEntities []goke.Entity
	seenPairs        map[uint64]struct{} // dedup Contactów per (EntityA.Index, EntityB.Index) - klucz: idxA<<32 | idxB
}

func NewCollisionSystem(resouces *Resources) goke.System {
	return &CollisionSystem{
		Resources: resouces,
	}
}

func (s *CollisionSystem) Init(ecs *goke.ECS) {
	s.collisionView = goke.NewView3[Position, Velocity, Collision](ecs)
	s.seenPairs = make(map[uint64]struct{}, 256)
}

func (s *CollisionSystem) Update(lookup goke.Lookup, sched *goke.Schedule, d time.Duration) {
	const solverIterations = 16
	const probeExpandMaring = 32
	s.contactsBuffer = s.contactsBuffer[:0]
	s.contactsEntities = s.contactsEntities[:0]
	clear(s.seenPairs)
	s.broadPhase(sched, probeExpandMaring)

	s.narrowPhase(solverIterations)
	s.space.Flush(func(spatial.AABB) {})
}

func (s *CollisionSystem) broadPhase(sched *goke.Schedule, probeExpandMargin uint32) {
	for page := range s.collisionView.All() {
		for i, entityA := range page.Entity {
			pos, vel, col := &page.Comp1[i], &page.Comp2[i], &page.Comp3[i]

			checkFunc := func(boxA geom.AABB[uint32], fragA plane.FragPosition) {
				s.space.Query(boxA, func(idB uint64, fragB plane.FragPosition) {
					entityB := goke.Entity(idB)

					if entityA.Index() >= entityB.Index() {
						return
					}
					// Dedup: dla pary (A, B) z fragmentami Query może trafiać kilka kombinacji
					// (np. A.MAIN-B.MAIN, A.MAIN-B.FRAG_RIGHT, A.FRAG_RIGHT-B.MAIN, ...).
					// Bez tej deduplikacji powstaje kilka Contactów dla jednej pary entities,
					// co prowadzi do wielokrotnego swap velocity = parzysta liczba swapów = no change.
					key := uint64(entityA.Index())<<32 | uint64(entityB.Index())
					if _, exists := s.seenPairs[key]; exists {
						return
					}
					s.seenPairs[key] = struct{}{}

					s.contactsEntities = append(s.contactsEntities, entityB)
					s.contactsBuffer = append(s.contactsBuffer, Contact{
						EntityA: entityA, EntityB: entityB,
						PosA:  pos,
						VelA:  vel,
						ColA:  col,
						FragA: fragA,
						FragB: fragB,
					})
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
	}
	return
}

func (s *CollisionSystem) narrowPhase(solverIterations int) {
	now := time.Now()

	for i, item := range s.collisionView.Filter(s.contactsEntities) {
		s.contactsBuffer[i].PosB = item.Comp1
		s.contactsBuffer[i].VelB = item.Comp2
		s.contactsBuffer[i].ColB = item.Comp3
	}

	for iter := 0; iter < solverIterations; iter++ {
		for i := range s.contactsBuffer {
			contact := &s.contactsBuffer[i]

			boxA, boxB, penetrationVec := contact.findActiveCollision()

			if penetrationVec.X == 0 && penetrationVec.Y == 0 {
				continue
			}

			if !contact.resolved {
				// ta funkcja zmienia Vel (dla A i B) (componenty z ECS!)
				contact.resolveVelocity(penetrationVec)
				contact.updateStats(now)
				s.collisionCounter++
				contact.resolved = true
			}

			if mtv1, mtv2, ok := contact.calculateMtv(boxA, boxB, false); ok == true {
				// tu zmienia Pos (dla A i B) (componenty z ECS!)
				s.space.Translate(uint64(contact.EntityA), &contact.PosA.AABB, mtv1)
				s.space.Translate(uint64(contact.EntityB), &contact.PosB.AABB, mtv2)
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

// findActiveCollision iteruje przez wszystkie kombinacje (mainBoxA + fragmenty A) ×
// (mainBoxB + fragmenty B) i zwraca tę kombinację boxów wraz z penetracją, która ma
// największą "siłę kolizji" (|pen.X| + |pen.Y|). Jeśli żadna kombinacja nie ma overlapu,
// zwraca zerowe wartości.
//
// Po push-apart między iteracjami solwera geometria może się zmienić: fragment użyty
// pierwotnie w broadphase znika (bo obiekt już nie wystaje za krawędź), ale obiekty
// wciąż mogą się przecinać przez mainBoxy lub inne kombinacje fragmentów. Statyczne
// freshBoxA/freshBoxB zamrażały oryginalny FragA/FragB → traciły takie zmiany.
func (c *Contact) findActiveCollision() (boxA, boxB geom.AABB[uint32], pen geom.Vec[int32]) {
	bestArea := int32(0)

	visitA := func(bA geom.AABB[uint32]) {
		visitB := func(bB geom.AABB[uint32]) {
			p := c.penetration(bA, bB)
			if p.X == 0 && p.Y == 0 {
				return
			}
			area := abs32(p.X) + abs32(p.Y)
			if area > bestArea {
				bestArea = area
				boxA = bA
				boxB = bB
				pen = p
			}
		}
		// B.MAIN
		visitB(c.PosB.AABB.AABB)
		// B fragmenty (jeśli są)
		c.PosB.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
			visitB(b)
			return true
		})
	}

	// A.MAIN
	visitA(c.PosA.AABB.AABB)
	// A fragmenty (jeśli są)
	c.PosA.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
		visitA(b)
		return true
	})

	return
}

func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
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
