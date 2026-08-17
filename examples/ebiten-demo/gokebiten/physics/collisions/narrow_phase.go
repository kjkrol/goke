package collisions

import (
	"time"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg"
	"github.com/kjkrol/gokg/spatial"
	"github.com/kjkrol/uid"
)

var _ goke.System = (*NarrowPhase)(nil)

type NarrowPhase struct {
	space       *gokg.Space
	handler     CollisionHandler
	hitDuration time.Duration

	hitQry    *goke.Query
	pos       goke.Comp[kinematics.Position]
	vel       goke.Comp[kinematics.Velocity]
	hitTag    goke.Comp[Hit]
	collision goke.Comp[Collision]

	dynQry *goke.Query
	dynPos goke.Comp[kinematics.Position]
	dynVel goke.Comp[kinematics.Velocity]

	staticQuery *goke.Query
	staticPos   goke.Comp[kinematics.Position]

	sensorIDs map[uid.UID64]struct{}

	removeEditor *goke.Editor

	contactsBuffer []Contact

	dynSeeded, staticSeeded bool

	confirmed map[uid.UID64]bool
}

func NewNarrowPhase(space *gokg.Space, handler CollisionHandler, hitDuration time.Duration) *NarrowPhase {
	return &NarrowPhase{
		space: space, handler: handler, hitDuration: hitDuration,
		confirmed: make(map[uid.UID64]bool),
		sensorIDs: make(map[uid.UID64]struct{}),
	}
}

func (n *NarrowPhase) Init(si *goke.SysInit) {
	n.hitQry = si.NewQueryBuilder(&n.pos, &n.vel, &n.hitTag, &n.collision).Build()
	n.dynQry = si.NewQueryBuilder(&n.dynPos, &n.dynVel).Build()
	n.staticQuery = si.NewQueryBuilder(&n.staticPos).Exclude(goke.Exclude[kinematics.Velocity]()).Build()
	if init, ok := n.handler.(Initializer); ok {
		init.Init(si)
	}
	n.removeEditor = n.hitQry.NewEditorBuilder().Remove(goke.Remove[Hit]()).Build()
	sensorQry := si.NewQueryBuilder().Include(goke.Include[Sensor]()).Build()
	sensorQry.All()
	for sensorQry.Next() {
		for _, id := range sensorQry.Cursor().IDs {
			n.sensorIDs[id] = struct{}{}
		}
	}
}

func (n *NarrowPhase) Update(cb *goke.CmdBuf, _ time.Duration) {
	const solverIterations = 16
	n.contactsBuffer = n.contactsBuffer[:0]
	clear(n.confirmed)

	n.pair()
	n.solve(cb, solverIterations)
	n.finalizeHitTags(cb)
}

func (n *NarrowPhase) isSensor(id uid.UID64) bool {
	_, ok := n.sensorIDs[id]
	return ok
}

func (n *NarrowPhase) pair() {
	n.dynSeeded, n.staticSeeded = false, false

	n.hitQry.All()
	for n.hitQry.Next() {
		cursor := n.hitQry.Cursor()
		posSlice := n.pos.Slice(cursor)
		velSlice := n.vel.Slice(cursor)
		collisionSlice := n.collision.Slice(cursor)
		for i, entityA := range cursor.IDs {
			p, v, c := &posSlice[i], &velSlice[i], &collisionSlice[i]
			n.confirmed[entityA] = false

			for ti := uint8(0); ti < c.TouchingCount; ti++ {
				entityB := c.Touching[ti]
				if entityA.Index() >= entityB.Index() {
					continue
				}
				posB, velB, ok := n.resolveB(entityB)
				if !ok {
					continue
				}
				n.contactsBuffer = append(n.contactsBuffer, Contact{
					A: contactSide{Entity: entityA, Pos: p, Vel: v, IsSensor: n.isSensor(entityA)},
					B: contactSide{Entity: entityB, Pos: posB, Vel: velB, IsSensor: n.isSensor(entityB)},
				})
			}

			c.clear()
		}
	}
}

func (n *NarrowPhase) resolveB(id uid.UID64) (pos *kinematics.Position, vel *kinematics.Velocity, ok bool) {
	ok = n.dynSeeded && n.dynQry.SeekH(id)
	if !ok {
		ok = n.dynQry.Seek(id)
		n.dynSeeded = ok
	}
	if ok {
		cur := n.dynQry.Cursor()
		return n.dynPos.At(cur), n.dynVel.At(cur), true
	}

	ok = n.staticSeeded && n.staticQuery.SeekH(id)
	if !ok {
		ok = n.staticQuery.Seek(id)
		n.staticSeeded = ok
	}
	if !ok {
		return nil, nil, false
	}
	return n.staticPos.At(n.staticQuery.Cursor()), nil, true
}

func (n *NarrowPhase) solve(cb *goke.CmdBuf, solverIterations int) {
	for range solverIterations {
		for i := range n.contactsBuffer {
			contact := &n.contactsBuffer[i]

			boxA, boxB, penetrationVec := contact.findActiveCollision()
			if penetrationVec.X == 0 && penetrationVec.Y == 0 {
				continue
			}

			isSensorContact := contact.A.IsSensor || contact.B.IsSensor

			if !contact.resolved {
				if !isSensorContact && n.handler != nil {
					n.handler.OnCollision(cb, CollisionEvent{
						EntityA: contact.A.Entity, EntityB: contact.B.Entity,
						PosA: contact.A.Pos, PosB: contact.B.Pos,
						VelA: contact.A.Vel, VelB: contact.B.Vel,
						Penetration: penetrationVec,
					})
				}
				if _, ok := n.confirmed[contact.A.Entity]; ok {
					n.confirmed[contact.A.Entity] = true
				}
				if _, ok := n.confirmed[contact.B.Entity]; ok {
					n.confirmed[contact.B.Entity] = true
				}
				contact.resolved = true
			}

			if isSensorContact {
				continue
			}

			isStaticB := contact.B.Vel == nil
			if mtv1, mtv2, ok := contact.calculateMtv(boxA, boxB, isStaticB); ok {
				n.space.Translate(uint64(contact.A.Entity), &contact.A.Pos.AABB, mtv1)
				if !isStaticB {
					n.space.Translate(uint64(contact.B.Entity), &contact.B.Pos.AABB, mtv2)
				}
			}
		}
	}
	n.space.Flush(func(spatial.AABB) {})
}

func (n *NarrowPhase) finalizeHitTags(cb *goke.CmdBuf) {
	until := time.Now().Add(n.hitDuration)

	n.hitQry.All()
	for n.hitQry.Next() {
		cursor := n.hitQry.Cursor()
		tags := n.hitTag.Slice(cursor)
		buf := n.hitQry.BeginMigrate(cb)
		for i, id := range cursor.IDs {
			confirmedThisTick, tracked := n.confirmed[id]
			if !tracked {
				continue
			}
			if confirmedThisTick {
				tags[i].ExpiresAt = until
				continue
			}
			if !tags[i].ExpiresAt.IsZero() {
				continue
			}
			buf.Add(id)
		}
		buf.Commit(n.removeEditor)
	}
}
