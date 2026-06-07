package main

import (
	"image/color"
	"math"
	"math/rand/v2"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
)

type EntitiesInitSystem struct {
	*Resources
	blueprint *goke.Blueprint4[Position, Velocity, Collision, Appearance]
}

var _ goke.System = (*EntitiesInitSystem)(nil)

func NewEntitiesInitSystem(reources *Resources) goke.System {
	return &EntitiesInitSystem{
		Resources: reources,
	}
}

func (s *EntitiesInitSystem) Init(ecs *goke.ECS) {
	spawnEntitiesNumber := s.entityCounter
	s.blueprint = goke.NewBlueprint4[Position, Velocity, Collision, Appearance](ecs)

	gridSize := math.Ceil(math.Sqrt(float64(spawnEntitiesNumber)))
	cols := uint32(gridSize)

	// 2. Calculate dynamic spacing to fill the whole ScreenWidth/Height
	cellWidth := s.space.Width / cols
	cellHeight := s.space.Height / cols

	index := 0
	for page := range s.blueprint.Create(spawnEntitiesNumber) {
		for j, entity := range page.Entity {
			position, velocity, collision, appearance := &page.Comp1[j], &page.Comp2[j], &page.Comp3[j], &page.Comp4[j]
			row := uint32(index) / cols
			col := uint32(index) % cols

			spawnEntity(entity, position, velocity, collision, appearance,
				cellWidth, cellHeight, row, col)

			s.space.Insert(uint64(entity.Index()), position.AABB)
			index++
		}
	}
	s.space.Flush(nil)
}

func (s *EntitiesInitSystem) Update(goke.Lookup, *goke.Schedule, time.Duration) {}

func spawnEntity(
	entity goke.Entity,
	position *Position,
	velocity *Velocity,
	collistion *Collision,
	appearance *Appearance,
	cellWidth, cellHeight uint32,
	row, col uint32,
) {
	// 3. Center the entity within its allocated cell
	// Cell center minus half of RectSize
	startX := (col * cellWidth) + (cellWidth / 2) - (RectSize / 2)
	startY := (row * cellHeight) + (cellHeight / 2) - (RectSize / 2)

	startPos := geom.NewVec(startX, startY)
	aabb := plane.NewAABB(startPos, RectSize, RectSize)

	*position = Position{
		AABB: aabb,
		// accX: 0,
		// accY: 0,
	}

	// Velocity initialization
	dx := rand.Int32N(401) - 200
	dy := rand.Int32N(401) - 200

	if dx >= 0 && dx < 50 {
		dx = 10
	} else if dx < 0 && dx > -50 {
		dx = -10
	}

	*velocity = Velocity{
		Vec: geom.NewVec(dx, dy),
	}

	*collistion = Collision{}

	*appearance = Appearance{
		Color: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
}
