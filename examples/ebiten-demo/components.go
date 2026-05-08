package main

import (
	"image/color"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/plane"
)

type Position struct {
	plane.AABB[uint32]
	// Accumulators for sub-pixel movement
	accX float64
	accY float64
}
type Velocity struct{ geom.Vec[int32] }

type Appearance struct {
	Color color.RGBA
}

var (
	posDesc goke.ComponentDesc
	velDesc goke.ComponentDesc
	appDesc goke.ComponentDesc
)

func InitBaseComponents(ecs *goke.ECS) {
	posDesc = goke.RegisterComponent[Position](ecs)
	velDesc = goke.RegisterComponent[Velocity](ecs)
	appDesc = goke.RegisterComponent[Appearance](ecs)
}
