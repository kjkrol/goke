package main

import (
	"image/color"
	"math"
	"math/rand/v2"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/plane"
	"github.com/kjkrol/gokg/pkg/spatial"
)

type EntitiesInitSystem struct {
	*Resource
	blueprint *goke.Blueprint3[Position, Velocity, Appearance]
}

var _ goke.System = (*EntitiesInitSystem)(nil)

func NewEntitiesInitSystem(reource *Resource) *EntitiesInitSystem {
	return &EntitiesInitSystem{
		Resource: reource,
	}
}

func (s *EntitiesInitSystem) Init(ecs *goke.ECS) {
	spawnEntitiesNumber := s.entityCount
	s.blueprint = goke.NewBlueprint3[Position, Velocity, Appearance](ecs)

	gridSize := math.Ceil(math.Sqrt(float64(spawnEntitiesNumber)))
	cols := uint32(gridSize)

	// 2. Calculate dynamic spacing to fill the whole ScreenWidth/Height
	cellWidth := uint32(s.grid.witdh / cols)
	cellHeight := uint32(s.grid.height / cols)

	for i := 0; i < spawnEntitiesNumber; i++ {
		entity, position, velocity, appearance := s.blueprint.Create()

		row := uint32(i) / cols
		col := uint32(i) % cols

		spawnEntity(entity, position, velocity, appearance,
			cellWidth, cellHeight, row, col)

		aabb := position.AABB
		s.grid.spatialIndex.QueueInsert(uint64(entity.Index()), aabb.AABB)
	}
	s.grid.spatialIndex.Flush(func(a spatial.AABB) {})
}

func (s *EntitiesInitSystem) Update(goke.Lookup, *goke.Schedule, time.Duration) {}

func spawnEntity(
	entity goke.Entity,
	position *Position,
	velocity *Velocity,
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

	*appearance = Appearance{
		Color: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
}
