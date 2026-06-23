package main

import (
	"image/color"
	"time"

	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
)

type Position struct {
	plane.AABB[uint32]
	// Accumulators for sub-pixel movement
	accX float64
	accY float64
}
type Velocity struct{ geom.Vec[int32] }

type Collision struct {
	timestamp time.Time
	counter   uint8
}

type Appearance struct {
	Color    color.RGBA
	SpriteID uint8
}

const SpriteCount = 4
