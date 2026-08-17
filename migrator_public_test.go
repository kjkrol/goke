package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// TestMigratorBuilder_AddComponent exercises the public Migrator API end to
// end: NewMigratorBuilder(comps...).Build(), applied via Query.BeginMigrate
// inside a system's Update, applied at Sync.
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

	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			buf := query.BeginMigrate(cb)
			for _, id := range cursor.IDs {
				buf.Add(id)
			}
			buf.Commit(addVel)
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

// TestMigratorBuilder_RemoveComponent exercises the Remove side: a Migrator
// built via NewMigratorBuilder().Remove(...).Build() removes a component
// from every matched entity.
func TestMigratorBuilder_RemoveComponent(t *testing.T) {
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
	removeVel := ecs.NewMigratorBuilder().Remove(goke.Remove[Velocity]()).Build()

	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			buf := query.BeginMigrate(cb)
			for _, id := range cursor.IDs {
				buf.Add(id)
			}
			buf.Commit(removeVel)
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
