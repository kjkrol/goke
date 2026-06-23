package main

import (
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
	"github.com/kjkrol/gokg/spatial"
	"github.com/kjkrol/uid"
)

var _ goke.System = (*CollisionSystem)(nil)

type CollisionSystem struct {
	*Resources
	collisionQuery   *goke.Query
	pos              goke.Comp[Position]
	vel              goke.Comp[Velocity]
	coll             goke.Comp[Collision]
	contactsBuffer   []Contact
	contactsEntities []uid.UID64
	// seenPairs        map[uint64]struct{} // dedup Contactów per (EntityA.Index, EntityB.Index) - klucz: idxA<<32 | idxB

}

func NewCollisionSystem(resouces *Resources) goke.System {
	return &CollisionSystem{
		Resources: resouces,
	}
}

func (s *CollisionSystem) Init(ecs *goke.ECS) {
	s.collisionQuery = ecs.NewQueryBuilder(&s.pos, &s.vel, &s.coll).Build()
	// s.seenPairs = make(map[uint64]struct{}, 256)
}

func (s *CollisionSystem) Update(sched *goke.CmdBuf, d time.Duration) {
	const solverIterations = 16
	const probeExpandMaring = 32
	s.contactsBuffer = s.contactsBuffer[:0]
	s.contactsEntities = s.contactsEntities[:0]
	// clear(s.seenPairs)
	s.broadPhase(probeExpandMaring)

	s.narrowPhase(solverIterations)
	s.space.Flush(func(spatial.AABB) {})
}

func (s *CollisionSystem) broadPhase(probeExpandMargin uint32) {
	s.collisionQuery.All()
	for s.collisionQuery.Next() {
		posSlice := s.pos.Slice(&s.collisionQuery.Cursor)
		velSlice := s.vel.Slice(&s.collisionQuery.Cursor)
		collSlice := s.coll.Slice(&s.collisionQuery.Cursor)
		for i, entityA := range s.collisionQuery.Cursor.IDs {
			p, v, c := &posSlice[i], &velSlice[i], &collSlice[i]

			checkFunc := func(boxA geom.AABB[uint32], fragA plane.FragPosition) {
				s.space.Query(boxA, func(idB uint64, fragB plane.FragPosition) {
					entityB := uid.UID64(idB)

					if entityA.Index() >= entityB.Index() {
						return
					}
					s.contactsEntities = append(s.contactsEntities, entityB)
					s.contactsBuffer = append(s.contactsBuffer, Contact{
						EntityA: entityA, EntityB: entityB,
						PosA:  p,
						VelA:  v,
						ColA:  c,
						FragA: fragA,
						FragB: fragB,
					})
				})
			}

			probeBoxA := p.AABB
			s.space.ExpandOnly(&probeBoxA, probeExpandMargin)
			checkFunc(probeBoxA.AABB, plane.FRAG_MAIN)
			probeBoxA.VisitFragments(func(fragA plane.FragPosition, boxA geom.AABB[uint32]) bool {
				checkFunc(boxA, fragA)
				return true
			})
		}
	}
}

func (s *CollisionSystem) narrowPhase(solverIterations int) {
	now := time.Now()

	s.collisionQuery.Pick(s.contactsEntities)
	for s.collisionQuery.Next() {
		s.contactsBuffer[s.collisionQuery.Idx].PosB = s.pos.At(&s.collisionQuery.Cursor)
		s.contactsBuffer[s.collisionQuery.Idx].VelB = s.vel.At(&s.collisionQuery.Cursor)
		s.contactsBuffer[s.collisionQuery.Idx].ColB = s.coll.At(&s.collisionQuery.Cursor)
	}

	for range solverIterations {
		for i := range s.contactsBuffer {
			contact := &s.contactsBuffer[i]

			boxA, boxB, penetrationVec := contact.findActiveCollision()

			if penetrationVec.X == 0 && penetrationVec.Y == 0 {
				continue
			}

			if !contact.resolved {
				// ta funkcja zmienia Vel (dla A i B) (componenty z ECS!)
				updateVelocity(contact.VelA, contact.VelB, penetrationVec)
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
	EntityA uid.UID64
	EntityB uid.UID64
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
//
// Optymalizacja: fast-path dla typowego przypadku (oba obiekty bez fragmentów) -
// 1 wywołanie penetration zamiast 16. Slow-path iteruje tylko te kombinacje które
// realnie mogą wystąpić.
func (c *Contact) findActiveCollision() (boxA, boxB geom.AABB[uint32], pen geom.Vec[int32]) {
	mainA := c.PosA.AABB.AABB
	mainB := c.PosB.AABB.AABB

	hasFragsA := hasAnyFragment(&c.PosA.AABB)
	hasFragsB := hasAnyFragment(&c.PosB.AABB)

	// FAST PATH: brak fragmentów po obu stronach - tylko mainBox-mainBox możliwy.
	if !hasFragsA && !hasFragsB {
		boxA = mainA
		boxB = mainB
		pen = c.penetration(mainA, mainB)
		return
	}

	// SLOW PATH: iteruj po wszystkich kombinacjach które realnie istnieją.
	bestArea := int32(0)

	tryCombo := func(bA, bB geom.AABB[uint32]) {
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

	// mainA × mainB
	tryCombo(mainA, mainB)

	// mainA × fragsB
	if hasFragsB {
		c.PosB.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
			tryCombo(mainA, b)
			return true
		})
	}

	// fragsA × mainB
	if hasFragsA {
		c.PosA.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
			tryCombo(b, mainB)
			return true
		})
	}

	// not needed
	// fragsA × fragsB
	// if hasFragsA && hasFragsB {
	// 	c.PosA.AABB.VisitFragments(func(_ plane.FragPosition, bA geom.AABB[uint32]) bool {
	// 		c.PosB.AABB.VisitFragments(func(_ plane.FragPosition, bB geom.AABB[uint32]) bool {
	// 			tryCombo(bA, bB)
	// 			return true
	// 		})
	// 		return true
	// 	})
	// }

	return
}

// hasAnyFragment sprawdza czy AABB ma jakiekolwiek aktywne fragmenty po wrap-around.
// Wykorzystuje VisitFragments z early-out (zwracając false z pierwszego callbacka).
func hasAnyFragment(ab *plane.AABB[uint32]) bool {
	has := false
	ab.VisitFragments(func(_ plane.FragPosition, _ geom.AABB[uint32]) bool {
		has = true
		return false
	})
	return has
}

func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// updateVelocity przetwarza fizykę odbicia AABB
func updateVelocity(velA, velB *Velocity, pen geom.Vec[int32]) {
	if pen.X != 0 {
		relVelX := int32(velA.Vec.X) - int32(velB.Vec.X)

		if (pen.X > 0 && relVelX > 0) || (pen.X < 0 && relVelX < 0) {
			return
		}

		tempX := velA.Vec.X
		velA.Vec.X = velB.Vec.X
		velB.Vec.X = tempX
	} else if pen.Y != 0 {
		relVelY := int32(velA.Vec.Y) - int32(velB.Vec.Y)

		if (pen.Y > 0 && relVelY > 0) || (pen.Y < 0 && relVelY < 0) {
			return
		}

		tempY := velA.Vec.Y
		velA.Vec.Y = velB.Vec.Y
		velB.Vec.Y = tempY
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
