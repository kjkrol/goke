package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// TestMigratorBuilder_AddComponent exercises the public Migrator API end to
// end: NewMigratorBuilder(comps...).Build(), CmdBufMassMigrate called from
// inside a system's Update using Query.ChunkSnapshot, applied at Sync.
func TestMigratorBuilder_AddComponent(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(2)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	var vel goke.Comp[Velocity]
	query := ecs.NewQueryBuilder(&pos).Exclude(goke.Exclude[Velocity]()).Build()
	addVel := ecs.NewMigratorBuilder(&vel).Build()

	var matched []uid.UID64
	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			matched = matched[:0]
			matched = append(matched, cursor.IDs...)
			if len(matched) == 0 {
				continue
			}
			snap := query.ChunkSnapshot()
			goke.CmdBufMassMigrate(cb, addVel, snap, matched)
		}
	})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if !hasComp[Velocity](ecs, id) {
			t.Errorf("entity %v: expected Velocity added via Migrator", id)
		}
	}
}

// TestMigratorBuilder_DeleteComponent exercises the Delete side: a Migrator
// built via NewMigratorBuilder().Delete(...).Build() removes a component
// from every matched entity.
func TestMigratorBuilder_DeleteComponent(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	factory := ecs.NewFactory(&pos, &vel)
	factory.Create(2)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	query := ecs.NewQueryBuilder(&pos, &vel).Build()
	removeVel := ecs.NewMigratorBuilder().Delete(goke.Del[Velocity]()).Build()

	var matched []uid.UID64
	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			matched = matched[:0]
			matched = append(matched, cursor.IDs...)
			if len(matched) == 0 {
				continue
			}
			snap := query.ChunkSnapshot()
			goke.CmdBufMassMigrate(cb, removeVel, snap, matched)
		}
	})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if hasComp[Velocity](ecs, id) {
			t.Errorf("entity %v: expected Velocity removed via Migrator", id)
		}
		if !hasComp[Position](ecs, id) {
			t.Errorf("entity %v: expected Position to remain", id)
		}
	}
}
