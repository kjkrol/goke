package collisions

import (
	"time"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
	"github.com/kjkrol/uid"
)

var _ goke.System = (*BroadPhase)(nil)

type BroadPhase struct {
	space       *gokg.Space
	query       *goke.Query
	pos         goke.Comp[kinematics.Position]
	vel         goke.Comp[kinematics.Velocity]
	hit         goke.Comp[Hit]
	collision   goke.Comp[Collision]
	addMigrator *goke.Migrator
}

func NewBroadPhase(space *gokg.Space) *BroadPhase {
	return &BroadPhase{space: space}
}

func (b *BroadPhase) Init(ecs *goke.ECS) {
	b.query = ecs.NewQueryBuilder(&b.pos, &b.vel, &b.collision).Build()
	goke.RegComp[Hit](ecs)
	b.addMigrator = ecs.NewMigratorBuilder(&b.hit).Build()
}

func (b *BroadPhase) Update(cb *goke.CmdBuf, _ time.Duration) {
	const probeExpandMargin = 32
	b.query.All()
	for b.query.Next() {
		cursor := b.query.Cursor()
		posSlice := b.pos.Slice(cursor)
		collisionSlice := b.collision.Slice(cursor)
		buf := b.query.BeginMigrate(cb)
		for i, entityA := range cursor.IDs {
			found := false
			p := &posSlice[i]
			c := &collisionSlice[i]

			checkFunc := func(boxA geom.AABB[uint32]) {
				b.space.Query(boxA, func(idB uint64, _ plane.FragPosition) {
					entityB := uid.UID64(idB)
					if entityA == entityB {
						return
					}
					c.addTouching(entityB)
					if !found {
						buf.Add(entityA)
						found = true
					}
				})
			}

			probeBoxA := p.AABB
			b.space.ExpandOnly(&probeBoxA, probeExpandMargin)
			checkFunc(probeBoxA.AABB)
			probeBoxA.VisitFragments(func(_ plane.FragPosition, boxA geom.AABB[uint32]) bool {
				checkFunc(boxA)
				return true
			})
		}
		buf.Commit(b.addMigrator)
	}
}
