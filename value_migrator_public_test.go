package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// TestValueMigratorBuilder_WritesPerEntityValue exercises the public
// ValueMigrator API end to end: NewValueMigratorBuilder(comp).Build(),
// CmdBufAddCompValue called from inside a system's Update using
// Query.ChunkSnapshot, values written into the returned slice, applied at
// Sync — then read back afterward via a normal Query.
func TestValueMigratorBuilder_WritesPerEntityValue(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(3)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	var vel goke.Comp[Velocity]
	query := ecs.NewQueryBuilder(&pos).Exclude(goke.Exclude[Velocity]()).Build()
	vm := ecs.NewValueMigratorBuilder(&vel).Build()

	want := map[uid.UID64]Velocity{}
	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			snap := query.ChunkSnapshot()
			vals := goke.CmdBufAddCompValue(cb, vm, &vel, snap, cursor.IDs)
			for i, id := range cursor.IDs {
				v := Velocity{VX: float32(i) + 1, VY: float32(i) + 2}
				vals[i] = v
				want[id] = v
			}
		}
	})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	verifyQ := ecs.NewQueryBuilder(&pos, &vel).Build()
	verifyQ.All()
	got := map[uid.UID64]Velocity{}
	for verifyQ.Next() {
		cur := verifyQ.Cursor()
		vals := vel.Slice(cur)
		for i, id := range cur.IDs {
			got[id] = vals[i]
		}
	}

	for _, id := range ids {
		w, ok := want[id]
		if !ok {
			t.Fatalf("entity %v: no expected value recorded (test bug)", id)
		}
		g, ok := got[id]
		if !ok {
			t.Fatalf("entity %v: Velocity not found after migration", id)
		}
		if g != w {
			t.Errorf("entity %v: Velocity = %+v; want %+v", id, g, w)
		}
	}
}

// TestValueMigratorBuilder_RemoveComponent exercises Remove: entities start
// with Position and Discount; the ValueMigrator adds Velocity (with a
// per-entity value) and removes Discount in the same migration.
func TestValueMigratorBuilder_RemoveComponent(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)
	_ = goke.RegComp[Discount](ecs)

	var pos goke.Comp[Position]
	var disc goke.Comp[Discount]
	factory := ecs.NewFactory(&pos, &disc)
	factory.Create(2)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	var vel goke.Comp[Velocity]
	query := ecs.NewQueryBuilder(&pos, &disc).Build()
	vm := ecs.NewValueMigratorBuilder(&vel).Remove(goke.Remove[Discount]()).Build()

	want := map[uid.UID64]Velocity{}
	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			snap := query.ChunkSnapshot()
			vals := goke.CmdBufAddCompValue(cb, vm, &vel, snap, cursor.IDs)
			for i, id := range cursor.IDs {
				v := Velocity{VX: float32(i) + 1, VY: float32(i) + 2}
				vals[i] = v
				want[id] = v
			}
		}
	})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if hasComp[Discount](ecs, id) {
			t.Errorf("entity %v: expected Discount removed via ValueMigrator", id)
		}
		if !hasComp[Velocity](ecs, id) {
			t.Errorf("entity %v: expected Velocity added via ValueMigrator", id)
		}
	}
}

// TestValueMigratorBuilder_MismatchedType_Panics confirms the runtime size
// guard in CmdBufAddCompValue catches a Comp[T] that doesn't match the
// ValueMigrator's own added component — the one case Go's type system can't
// catch here, since ValueMigrator is untyped at the Go level.
func TestValueMigratorBuilder_MismatchedType_Panics(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(1)
	for factory.Next() {
	}

	var vel goke.Comp[Velocity]
	vm := ecs.NewValueMigratorBuilder(&vel).Build()
	query := ecs.NewQueryBuilder(&pos).Build()

	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			// pos does not match vm's own added component (Velocity) — must panic.
			goke.CmdBufAddCompValue(cb, vm, &pos, query.ChunkSnapshot(), cursor.IDs)
		}
	})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from mismatched Comp[T], got none")
		}
	}()
	ecs.Tick(time.Millisecond)
}
