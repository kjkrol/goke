package main

import (
	"image/color"
	"math"
	"math/rand/v2"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
	"github.com/kjkrol/uid"
)

type EntitiesInitSystem struct {
	*Resources
	factory *goke.Factory4[Position, Velocity, Collision, Appearance]
}

var _ goke.System = (*EntitiesInitSystem)(nil)

func NewEntitiesInitSystem(reources *Resources) goke.System {
	return &EntitiesInitSystem{
		Resources: reources,
	}
}

func (s *EntitiesInitSystem) Init(ecs *goke.ECS) {
	spawnEntitiesNumber := s.entityCounter
	s.factory = goke.NewFactory4[Position, Velocity, Collision, Appearance](ecs)

	gridSize := math.Ceil(math.Sqrt(float64(spawnEntitiesNumber)))
	cols := uint32(gridSize)

	// 2. Calculate dynamic spacing to fill the whole ScreenWidth/Height
	cellWidth := s.space.Width / cols
	cellHeight := s.space.Height / cols

	index := 0
	for chunk := range s.factory.Create(spawnEntitiesNumber) {
		for j, entityID := range chunk.Entity {
			position, velocity, collision, appearance := &chunk.Comp1[j], &chunk.Comp2[j], &chunk.Comp3[j], &chunk.Comp4[j]
			row := uint32(index) / cols
			col := uint32(index) % cols

			spawnEntity(entityID, position, velocity, collision, appearance,
				cellWidth, cellHeight, row, col)

			s.space.Insert(uint64(entityID), position.AABB)
			index++
		}
	}
	s.space.Flush(nil)
}

func (s *EntitiesInitSystem) Update(goke.Lookup, *goke.CmdBuf, time.Duration) {}

func spawnEntity(
	entityID uid.UID64,
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
		Color:    color.RGBA{R: uint8(rand.IntN(206) + 50), G: uint8(rand.IntN(206) + 50), B: uint8(rand.IntN(206) + 50), A: 255},
		SpriteID: uint8(rand.IntN(SpriteCount)),
	}
}
