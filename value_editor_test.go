package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// TestValueEditorBuilder_WritesPerEntityValue exercises the public
// ValueEditor API end to end: query.NewValueEditorBuilder(comp).Build(),
// CmdBufAddCompValue called from inside a system's Update using
// Query.ChunkSnapshot, values written into the returned slice, applied at
// Sync — then read back afterward via a normal Query.
func TestValueEditorBuilder_WritesPerEntityValue(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var ids []uid.UID64
	var query *goke.Query
	var vm *goke.ValueEditor
	var verifyQ *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(3)
		for factory.Next() {
			ids = append(ids, factory.IDs...)
		}

		query = si.NewQueryBuilder(&pos).Exclude(goke.Exclude[Velocity]()).Build()
		vm = query.NewValueEditorBuilder(&vel).Build()
		verifyQ = si.NewQueryBuilder(&pos, &vel).Build()
	}})

	want := map[uid.UID64]Velocity{}
	sys := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
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
	}})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

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

// TestValueEditorBuilder_RemoveComponent exercises Remove: entities start
// with Position and Discount; the ValueEditor adds Velocity (with a
// per-entity value) and removes Discount in the same migration.
func TestValueEditorBuilder_RemoveComponent(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)
	_ = goke.RegComp[Discount](ecs)

	var pos goke.Comp[Position]
	var disc goke.Comp[Discount]
	var vel goke.Comp[Velocity]
	var ids []uid.UID64
	var query *goke.Query
	var vm *goke.ValueEditor
	var discQuery, velQuery *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos, &disc)
		factory.Create(2)
		for factory.Next() {
			ids = append(ids, factory.IDs...)
		}

		query = si.NewQueryBuilder(&pos, &disc).Build()
		vm = query.NewValueEditorBuilder(&vel).Remove(goke.Remove[Discount]()).Build()
		discQuery = si.NewQueryBuilder().Include(goke.Include[Discount]()).Build()
		velQuery = si.NewQueryBuilder().Include(goke.Include[Velocity]()).Build()
	}})

	want := map[uid.UID64]Velocity{}
	sys := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
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
	}})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if hasComp(discQuery, id) {
			t.Errorf("entity %v: expected Discount removed via ValueEditor", id)
		}
		if !hasComp(velQuery, id) {
			t.Errorf("entity %v: expected Velocity added via ValueEditor", id)
		}
	}
}

// TestValueEditorBuilder_MismatchedType_Panics confirms the runtime size
// guard in CmdBufAddCompValue catches a Comp[T] that doesn't match the
// ValueEditor's own added component — the one case Go's type system can't
// catch here, since ValueEditor is untyped at the Go level.
func TestValueEditorBuilder_MismatchedType_Panics(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var vm *goke.ValueEditor
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(1)
		for factory.Next() {
		}

		query = si.NewQueryBuilder(&pos).Build()
		vm = query.NewValueEditorBuilder(&vel).Build()
	}})

	sys := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			// pos does not match vm's own added component (Velocity) — must panic.
			goke.CmdBufAddCompValue(cb, vm, &pos, query.ChunkSnapshot(), cursor.IDs)
		}
	}})
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
