package elastic

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg/geom"
)

var _ collisions.CollisionHandler = (*Handler)(nil)

// Handler is a ready-made collision reaction: elastic velocity swap along
// the penetration axis. Independent of any bookkeeping — compose it with
// strategies/stats if you also want hit tracking.
type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) OnCollision(_ *goke.CmdBuf, e collisions.CollisionEvent) {
	if e.VelA == nil || e.VelB == nil {
		return // one side is immovable — nothing to swap
	}
	swapVelocity(e.VelA, e.VelB, e.Penetration)
}

// swapVelocity swaps the velocity components along the penetration axis,
// unless the objects are already separating on that axis.
func swapVelocity(velA, velB *kinematics.Velocity, pen geom.Vec[int32]) {
	if pen.X != 0 {
		relVelX := int32(velA.Vec.X) - int32(velB.Vec.X)

		if (pen.X > 0 && relVelX > 0) || (pen.X < 0 && relVelX < 0) {
			return
		}

		velA.Vec.X, velB.Vec.X = velB.Vec.X, velA.Vec.X
	} else if pen.Y != 0 {
		relVelY := int32(velA.Vec.Y) - int32(velB.Vec.Y)

		if (pen.Y > 0 && relVelY > 0) || (pen.Y < 0 && relVelY < 0) {
			return
		}

		velA.Vec.Y, velB.Vec.Y = velB.Vec.Y, velA.Vec.Y
	}
}
