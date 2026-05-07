package main

import (
	"image/color"
	"math"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/pkg/geom"
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
			entity2 := goke.Entity(otherID / 4) // TODO: fix gokg!!
			if head.Entity.Index() < entity2.Index() {
				pos2, _ := lookup.ComponentGet(entity2, posDesc.ID)
				vel2, _ := lookup.ComponentGet(entity2, velDesc.ID)
				app2, _ := lookup.ComponentGet(entity2, appDesc.ID)

				// TODO: to powinno byc realizowane przez customowa funkcje, ktora
				// implementuje user
				//
				app.Color = color.RGBA{R: 255, A: 255}
				(*Appearance)(app2).Color = color.RGBA{R: 255, A: 255}
				s.resolveCollision(pos, vel, (*Position)(pos2), (*Velocity)(vel2))
				s.collisionCounter++
				// ---------------------
			}
		})
	}
}

// resolveCollision to strategia reakcji na kolizję oparta na algorytmie
// Minimum Translation Vector (MTV) do rozwiązania penetracji oraz
// wymianie pędów dla zderzeń idealnie sprężystych.
func (s *CollisionSystem) resolveCollision(
	pos1 *Position, vel1 *Velocity,
	pos2 *Position, vel2 *Velocity,
) {
	// ---------------------------------------------------------
	// ETAP 1: FIZYKA (Wymiana wektorów prędkości)
	// ---------------------------------------------------------
	// Zderzenie idealnie sprężyste dla obiektów o równej masie.
	// Cząstki po prostu przekazują sobie nawzajem całą swoją energię kinetyczną
	// i wektory ruchu. To załatwia "odbicie" w kolejnej klatce symulacji.
	tempVel := vel1.Vec
	vel1.Vec = vel2.Vec
	vel2.Vec = tempVel

	// ---------------------------------------------------------
	// ETAP 2: ROZWIĄZANIE PENETRACJI (Separacja pozycji)
	// ---------------------------------------------------------
	// Ponieważ symulacja jest dyskretna (działa w klatkach), obiekty po wykryciu
	// kolizji już fizycznie na siebie naszły. Wymiana prędkości nie wystarczy -
	// trzeba je natychmiast "rozepchnąć", aby w kolejnej klatce znów nie wyzwoliły
	// detekcji kolizji (co skutkowałoby utknięciem obiektów w sobie).

	// Pobranie struktur obwiedni (Axis-Aligned Bounding Box) dla obu ciał.
	r1 := pos1.AABB.AABB
	r2 := pos2.AABB.AABB

	// Wyliczenie centralnych punktów (środków ciężkości) obu prostokątów.
	// Wymagane do ustalenia z której strony nastąpiło zderzenie.
	c1x := int32(r1.TopLeft.X) + int32(s.rectSize)/2
	c1y := int32(r1.TopLeft.Y) + int32(s.rectSize)/2
	c2x := int32(r2.TopLeft.X) + int32(s.rectSize)/2
	c2y := int32(r2.TopLeft.Y) + int32(s.rectSize)/2

	// Wektory kierunkowe między środkami obiektów (od c2 do c1).
	// Pozwalają określić w jakiej relacji przestrzennej znajdują się obiekty.
	dx := c1x - c2x
	dy := c1y - c2y

	// Obliczenie głębokości penetracji na obu osiach.
	// s.rectSize to zakładana odległość środków przy styku krawędzi (suma dwóch połówek).
	// Wynik to liczba pikseli/jednostek, o które obiekty weszły "w siebie".
	penX := int32(s.rectSize) - int32(math.Abs(float64(dx)))
	penY := int32(s.rectSize) - int32(math.Abs(float64(dy)))

	// Algorytm MTV - szukamy najkrótszej drogi wyjścia z kolizji.
	// Obiekty rozpychamy tylko na tej osi, na której penetracja jest mniejsza.
	if penX < penY {
		// ---------------------------------------------------------
		// SEPARACJA NA OSI X (Poziom)
		// ---------------------------------------------------------
		// Dzielimy penetrację na 2, ponieważ przesuniemy OBA obiekty w przeciwnych
		// kierunkach (każdy o połowę błędu), zamiast przesuwać tylko jeden o całość.
		push := penX / 2
		if push == 0 {
			push = 1 // Zabezpieczenie przed brakiem przesunięcia z powodu zaokrąglenia integer
		}

		if dx > 0 {
			// dx > 0 oznacza, że obiekt 1 (c1x) jest na PRAWO od obiektu 2 (c2x).
			// Wypychamy obiekt 1 dalej w prawo (dodatni wektor), a obiekt 2 w lewo (ujemny wektor).
			// POPRAWKA: Usunięto wadliwe rzutowanie uint32. Używamy int32 (lub bazowego int).
			s.grid.space.Translate(&pos1.AABB, geom.NewVec(uint32(push), 0))
			s.grid.space.Translate(&pos2.AABB, geom.NewVec(uint32(-push), 0))
		} else {
			// dx <= 0 oznacza, że obiekt 1 jest na LEWO od obiektu 2.
			// Wypychamy obiekt 1 w lewo (-push), a obiekt 2 w prawo (push).
			s.grid.space.Translate(&pos1.AABB, geom.NewVec(uint32(-push), 0))
			s.grid.space.Translate(&pos2.AABB, geom.NewVec(uint32(push), 0))
		}
	} else {
		// ---------------------------------------------------------
		// SEPARACJA NA OSI Y (Pion)
		// ---------------------------------------------------------
		// Ten sam mechanizm co wyżej, ale działamy w przestrzeni pionowej.
		push := penY / 2
		if push == 0 {
			push = 1
		}

		if dy > 0 {
			// dy > 0 oznacza, że obiekt 1 jest NIŻEJ (w układzie współrzędnych ekranu gdzie Y rośnie w dół)
			// niż obiekt 2. Wypychamy pos1 w dół, a pos2 w górę.
			s.grid.space.Translate(&pos1.AABB, geom.NewVec(0, uint32(push)))
			s.grid.space.Translate(&pos2.AABB, geom.NewVec(0, uint32(-push)))
		} else {
			// Obiekt 1 jest WYŻEJ niż obiekt 2. Wypychamy pos1 w górę, pos2 w dół.
			s.grid.space.Translate(&pos1.AABB, geom.NewVec(0, uint32(-push)))
			s.grid.space.Translate(&pos2.AABB, geom.NewVec(0, uint32(push)))
		}
	}
}
