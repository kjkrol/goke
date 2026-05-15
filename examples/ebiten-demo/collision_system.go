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

// Struktury pomocnicze do zarządzania kontaktami
type entityPair struct {
	A goke.Entity
	B goke.Entity
}

type Contact struct {
	EntityA goke.Entity
	EntityB goke.Entity
	PosA    *Position
	PosB    *Position
	VelA    *Velocity
	VelB    *Velocity
	ColA    *Collision
	ColB    *Collision
}

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
	var contacts []Contact
	processedPairs := make(map[entityPair]bool)

	// ==========================================
	// 1. BROAD PHASE
	// ==========================================
	for head := range s.collisionView.All() {
		pos, vel, col := head.V1, head.V2, head.V3
		entityA := head.Entity

		if vel.X == 0 && vel.Y == 0 {
			continue
		}

		checkFunc := func(boxA geom.AABB[uint32]) {
			s.grid.spatialIndex.QueryRange(boxA, func(otherID uint64) {
				entityB := goke.Entity(otherID)

				if entityA.Index() < entityB.Index() {
					pair := entityPair{A: entityA, B: entityB}

					if processedPairs[pair] {
						return
					}
					processedPairs[pair] = true

					posBptr, _ := lookup.ComponentGet(entityB, s.compDescs.posDesc.ID)
					velBptr, _ := lookup.ComponentGet(entityB, s.compDescs.velDesc.ID)
					colBptr, _ := lookup.ComponentGet(entityB, s.compDescs.colDesc.ID)

					contacts = append(contacts, Contact{
						EntityA: entityA, EntityB: entityB,
						PosA: pos, PosB: (*Position)(posBptr),
						VelA: vel, VelB: (*Velocity)(velBptr),
						ColA: col, ColB: (*Collision)(colBptr),
					})
				}
			})
		}

		// Sprawdzamy główny AABB i fragmenty, przekazując je do checkFunc
		checkFunc(pos.AABB.AABB)
		pos.AABB.VisitFragments(func(_ plane.FragPosition, boxA geom.AABB[uint32]) bool {
			checkFunc(boxA)
			return true
		})
	}

	now := time.Now()
	s.processContacts(contacts, now)
	s.grid.spatialIndex.Flush(func(a spatial.AABB) {})
}

// processContacts zarządza głównym przepływem wąskiego wykrywania (Narrow Phase) i rozwiązywania (Resolution)
func (s *CollisionSystem) processContacts(contacts []Contact, now time.Time) {
	// Stan współdzielony podczas rozwiązywania klatki
	velocitiesResolved := make([]bool, len(contacts))
	entitiesToUpdate := make(map[goke.Entity]*Position)

	s.runIterativeSolver(contacts, velocitiesResolved, entitiesToUpdate, now)

	// Aktualizacja siatki przestrzennej na samym końcu, tylko dla obiektów, które fizycznie zmieniły pozycję
	for entity, pos := range entitiesToUpdate {
		s.grid.spatialIndex.QueueUpdate(uint64(entity), pos.AABB.AABB, true)
	}
}

// runIterativeSolver wykonuje pętle wypychania obiektów dla ustabilizowania fizyki
func (s *CollisionSystem) runIterativeSolver(contacts []Contact, velocitiesResolved []bool, entitiesToUpdate map[goke.Entity]*Position, now time.Time) {
	const solverIterations = 8

	for iter := 0; iter < solverIterations; iter++ {
		for i := range contacts {
			s.solveSingleContact(&contacts[i], i, velocitiesResolved, entitiesToUpdate, now)
		}
	}
}

// solveSingleContact przetwarza pojedynczy kontakt w danej iteracji solwera
func (s *CollisionSystem) solveSingleContact(c *Contact, index int, velocitiesResolved []bool, entitiesToUpdate map[goke.Entity]*Position, now time.Time) {
	bestBoxA, bestBoxB, finalPen, foundCollision := s.findCollisionFragments(c.PosA, c.PosB)

	// Jeśli brak kolizji w tej iteracji (obiekty mogły zostać rozsunięte w poprzednich krokach)
	if !foundCollision {
		return
	}

	// 1. Wypychanie (Resolution Phase - Pozycja, uwzględnia Velocity do wykrycia ścian)
	s.compensator.compensate(c.PosA, c.PosB, bestBoxA, bestBoxB, c.VelA, c.VelB)

	// Oznaczamy obiekty do aktualizacji w siatce (nadpisze się, jeśli obiekt uderza w wiele rzeczy)
	entitiesToUpdate[c.EntityA] = c.PosA
	entitiesToUpdate[c.EntityB] = c.PosB

	// 2. Prędkość i Eventy (wykonuje się TYLKO RAZ na zderzenie w klatce)
	if !velocitiesResolved[index] {
		c.ColA.timestamp = now
		c.ColA.counter++
		c.ColB.timestamp = now
		c.ColB.counter++
		s.collisionCounter++

		s.resolveVelocity(c.VelA, c.VelB, finalPen)

		velocitiesResolved[index] = true // Zabezpieczenie przed wielokrotnym rozliczeniem
	}
}

// findCollisionFragments szuka pary fragmentów (lub głównych AABB), która faktycznie na siebie nachodzi najmocniej
func (s *CollisionSystem) findCollisionFragments(posA, posB *Position) (geom.AABB[uint32], geom.AABB[uint32], geom.Vec[int32], bool) {
	var bestBoxA, bestBoxB geom.AABB[uint32]
	var finalPen geom.Vec[int32]
	var found bool

	// Szukamy najmniejszej drogi ucieczki
	var minMagnitude int32 = math.MaxInt32

	boxesA := []geom.AABB[uint32]{posA.AABB.AABB}
	posA.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
		boxesA = append(boxesA, b)
		return true
	})

	boxesB := []geom.AABB[uint32]{posB.AABB.AABB}
	posB.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
		boxesB = append(boxesB, b)
		return true
	})

	for _, bA := range boxesA {
		for _, bB := range boxesB {
			pen := s.compensator.penetration(bA, bB) // Zwraca wektor MTV dla tej pary fragmentów
			if pen.X != 0 || pen.Y != 0 {
				mag := int32(math.Abs(float64(pen.X)) + math.Abs(float64(pen.Y)))
				// Jeśli ta para fragmentów pozwala na KRÓTSZĄ ucieczkę, to jest nasz zwycięzca
				if mag < minMagnitude {
					minMagnitude = mag
					bestBoxA = bA
					bestBoxB = bB
					finalPen = pen
					found = true
				}
			}
		}
	}

	return bestBoxA, bestBoxB, finalPen, found
}

// resolveVelocity przetwarza fizykę odbicia AABB
// resolveVelocity przetwarza fizykę odbicia AABB
func (s *CollisionSystem) resolveVelocity(velA, velB *Velocity, pen geom.Vec[int32]) {
	if pen.X != 0 {
		relVelX := int32(velA.Vec.X) - int32(velB.Vec.X)

		// POPRAWKA: Jeśli wektor wypychania i wektor prędkości względnej
		// mają TEN SAM ZNAK, oznacza to, że obiekty JUŻ się od siebie oddalają.
		if (pen.X > 0 && relVelX > 0) || (pen.X < 0 && relVelX < 0) {
			return
		}

		tempX := velA.Vec.X
		velA.Vec.X = velB.Vec.X
		velB.Vec.X = tempX
	} else if pen.Y != 0 {
		relVelY := int32(velA.Vec.Y) - int32(velB.Vec.Y)

		// POPRAWKA: Ten sam znak = obiekty się oddalają.
		if (pen.Y > 0 && relVelY > 0) || (pen.Y < 0 && relVelY < 0) {
			return
		}

		tempY := velA.Vec.Y
		velA.Vec.Y = velB.Vec.Y
		velB.Vec.Y = tempY
	}
}

type penetrationCompensator struct {
	space plane.Space2D[uint32]
}

func (p penetrationCompensator) compensate(posA, posB *Position, boxA, boxB geom.AABB[uint32], velA, velB *Velocity) {
	// minimumTranslationVector teraz operuje na CZYSTYCH prostokątach i uwzględnia prędkość
	if mtv1, mtv2, ok := p.minimumTranslationVector(boxA, boxB, velA, velB); ok == true {
		// Ale przesuwanie wciąż operuje na całych obiektach dzięki Translate
		p.space.Translate(&posA.AABB, mtv1)
		p.space.Translate(&posB.AABB, mtv2)
	}
}

func (p penetrationCompensator) minimumTranslationVector(r1, r2 geom.AABB[uint32], velA, velB *Velocity) (mtv1, mtv2 geom.Vec[uint32], res bool) {
	pen := p.penetration(r1, r2) // Otrzymujemy gotowy, najkrótszy wektor z Narrow Phase

	if pen.X == 0 && pen.Y == 0 {
		return geom.Vec[uint32]{}, geom.Vec[uint32]{}, false
	}

	isStaticA := velA.Vec.X == 0 && velA.Vec.Y == 0
	isStaticB := velB.Vec.X == 0 && velB.Vec.Y == 0

	// Rozkłada wektor ucieczki na oba obiekty
	calculatePush := func(penetration int32) (int32, int32) {
		if isStaticA && !isStaticB {
			return 0, -penetration // A to ściana, wypchnij całe B w przeciwną stronę
		} else if !isStaticA && isStaticB {
			return penetration, 0 // B to ściana, wypchnij całe A
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

func (p *penetrationCompensator) penetration(r1, r2 geom.AABB[uint32]) geom.Vec[int32] {
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
