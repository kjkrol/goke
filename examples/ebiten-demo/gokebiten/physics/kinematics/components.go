package kinematics

import (
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
)

type Position struct {
	plane.AABB[uint32]
	// Sub-pixel movement accumulators, consumed by the movement integrator.
	AccX float64
	AccY float64
}

type Velocity struct{ geom.Vec[int32] }
