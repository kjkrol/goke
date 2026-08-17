package physics

import (
	"time"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions/strategies/debug"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg"
)

type Physics struct {
	space         *gokg.Space
	ecs           *goke.ECS
	minEntitySize uint32
	physicsStep   time.Duration

	handlers     []collisions.CollisionHandler
	extra        []goke.System
	extraHandles []goke.Runnable
	debug        bool
	hitDuration  time.Duration

	moveSystem  goke.Runnable
	broadPhase  goke.Runnable
	narrowPhase goke.Runnable
	built       bool
}

// New builds Physics for the given world and registers the kinematics
// system immediately — minEntitySize is the smallest entity size the world
// was provisioned for, used to cap per-tick displacement so nothing can
// tunnel through it undetected. Collision systems register lazily, on the
// first Run, since they must be built after entities exist.
func New(space *gokg.Space, ecs *goke.ECS, minEntitySize uint32, physicsStep time.Duration) *Physics {
	p := &Physics{space: space, ecs: ecs, minEntitySize: minEntitySize, physicsStep: physicsStep}
	p.moveSystem = p.useKinematics()
	return p
}

func (p *Physics) SetCollisionHandlers(handlers ...collisions.CollisionHandler) *Physics {
	p.handlers = handlers
	return p
}

func (p *Physics) DebugEnabled() *Physics {
	p.debug = true
	return p
}

func (p *Physics) SetHitExpires(duration time.Duration) *Physics {
	p.hitDuration = duration
	p.extra = append(p.extra, NewTagExpirySystem(func(h *collisions.Hit) time.Time { return h.ExpiresAt }))
	return p
}

func (p *Physics) RegSys(sys goke.System) *Physics {
	p.extra = append(p.extra, sys)
	return p
}

func (p *Physics) Run(ctx goke.RunCtx, d time.Duration) {
	if !p.built {
		p.build()
	}
	ctx.Run(p.moveSystem, d)
	ctx.Sync()

	ctx.Run(p.broadPhase, d)
	ctx.Sync()

	ctx.Run(p.narrowPhase, d)
	ctx.Sync()

	for _, h := range p.extraHandles {
		ctx.Run(h, d)
		ctx.Sync()
	}
}

func (p *Physics) build() {
	handlers := p.handlers
	if p.debug {
		handlers = append(handlers, debug.NewHandler())
	}
	broad, narrow := p.useCollisions(collisions.MultiHandler(handlers...), p.hitDuration)
	p.broadPhase = p.ecs.RegSys(broad)
	p.narrowPhase = p.ecs.RegSys(narrow)

	for _, sys := range p.extra {
		p.extraHandles = append(p.extraHandles, p.ecs.RegSys(sys))
	}
	p.built = true
}

func (p *Physics) useKinematics() goke.Runnable {
	// Cap displacement per tick to half the smallest entity's size, so
	// nothing can tunnel through it undetected.
	maxSpeed := int32(float64(p.minEntitySize) / 2 / p.physicsStep.Seconds())
	system := kinematics.NewSystem(p.space, maxSpeed)
	return p.ecs.RegSys(system)
}

func (p *Physics) useCollisions(handler collisions.CollisionHandler, hitDuration time.Duration) (*collisions.BroadPhase, *collisions.NarrowPhase) {
	return collisions.NewBroadPhase(p.space), collisions.NewNarrowPhase(p.space, handler, hitDuration)
}
