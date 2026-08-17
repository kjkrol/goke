package collisions

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/uid"
)

// CollisionEvent describes one newly-resolved contact — identity and geometry
// of the pair, valid only for the duration of the OnCollision call.
type CollisionEvent struct {
	EntityA, EntityB uid.UID64
	PosA, PosB       *kinematics.Position
	VelA, VelB       *kinematics.Velocity
	Penetration      geom.Vec[int32]
}

// CollisionHandler reacts to newly-resolved contacts; System calls it at most
// once per contact per tick, from within the narrow phase.
type CollisionHandler interface {
	OnCollision(cb *goke.CmdBuf, e CollisionEvent)
}

// Initializer is an optional extension of CollisionHandler: if the
// registered handler (or a handler composed via MultiHandler) implements it,
// System.Init calls it once at registration.
type Initializer interface {
	Init(*goke.SysInit)
}

// MultiHandler composes handlers to run in order within the same narrow-phase
// call — no extra passes over the contact buffer. Handlers share
// CollisionEvent's pointers, so later handlers observe mutations (e.g. a
// velocity response) made by earlier ones.
func MultiHandler(handlers ...CollisionHandler) CollisionHandler {
	return multiHandler(handlers)
}

var _ Initializer = multiHandler(nil)

type multiHandler []CollisionHandler

func (m multiHandler) OnCollision(cb *goke.CmdBuf, e CollisionEvent) {
	for _, h := range m {
		h.OnCollision(cb, e)
	}
}

func (m multiHandler) Init(si *goke.SysInit) {
	for _, h := range m {
		if init, ok := h.(Initializer); ok {
			init.Init(si)
		}
	}
}
