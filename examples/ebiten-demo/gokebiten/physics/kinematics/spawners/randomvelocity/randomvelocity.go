package randomvelocity

import (
	"math/rand/v2"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/gokg/geom"
)

var _ kinematics.Motion = (*RandomVelocity)(nil)

// RandomVelocity assigns a random velocity, X and Y independently drawn
// from [-Range, Range]; an X component whose magnitude falls below DeadZone
// is clamped up to MinSpeed so entities don't start nearly stationary.
type RandomVelocity struct {
	Range, DeadZone, MinSpeed int32
}

func New(rng, deadZone, minSpeed int32) *RandomVelocity {
	return &RandomVelocity{Range: rng, DeadZone: deadZone, MinSpeed: minSpeed}
}

func (m *RandomVelocity) InitialVelocity(index, count int) kinematics.Velocity {
	dx := rand.Int32N(2*m.Range+1) - m.Range
	dy := rand.Int32N(2*m.Range+1) - m.Range

	if dx >= 0 && dx < m.DeadZone {
		dx = m.MinSpeed
	} else if dx < 0 && dx > -m.DeadZone {
		dx = -m.MinSpeed
	}

	return kinematics.Velocity{Vec: geom.NewVec(dx, dy)}
}
