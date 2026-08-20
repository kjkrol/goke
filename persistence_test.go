package goke_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

func TestSaveLoad_RoundTrip(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Velocity]()

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos, &vel)
		factory.Create(3)
		for factory.Next() {
			positions := pos.Slice(&factory.Cursor)
			velocities := vel.Slice(&factory.Cursor)
			for i := range positions {
				positions[i] = Position{X: float32(i), Y: float32(i) * 10}
				velocities[i] = Velocity{VX: 1, VY: -1}
			}
		}
		ids = append(ids, factory.IDs...)
	}})

	path := filepath.Join(t.TempDir(), "save.bin")

	ecs.Pause()
	if err := ecs.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	ecs2 := goke.New()
	if err := ecs2.Load(path, goke.RegisterFor[Position](), goke.RegisterFor[Velocity]()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	var pos2 goke.Comp[Position]
	var vel2 goke.Comp[Velocity]
	var query *goke.Query
	ecs2.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		query = si.NewQueryBuilder(&pos2, &vel2).Build()
	}})

	for i, id := range ids {
		p := seekComp(query, &pos2, id)
		if p == nil {
			t.Fatalf("entity %d (%v): missing after Load", i, id)
		}
		if want := (Position{X: float32(i), Y: float32(i) * 10}); *p != want {
			t.Errorf("entity %d: Position = %+v, want %+v", i, *p, want)
		}
		v := seekComp(query, &vel2, id)
		if v == nil || *v != (Velocity{VX: 1, VY: -1}) {
			t.Errorf("entity %d: Velocity = %+v", i, v)
		}
	}
}

// kinematicsModule stands in for a self-contained, modular system that
// registers its own components — analogous to a gokebiten-style module the
// caller doesn't need to know the internals of.
type kinematicsModule struct {
	pos goke.Comp[Position]
	vel goke.Comp[Velocity]
}

func (m *kinematicsModule) Init(si *goke.SysInit) {
	si.ECS().RegComp[Position]()
	si.ECS().RegComp[Velocity]()
}

func (m *kinematicsModule) Update(cb *goke.CmdBuf, d time.Duration) {}

func (m *kinematicsModule) LoaderComps() []goke.LoaderComp {
	return []goke.LoaderComp{goke.RegisterFor[Position](), goke.RegisterFor[Velocity]()}
}

var (
	_ goke.System            = (*kinematicsModule)(nil)
	_ goke.ComponentProvider = (*kinematicsModule)(nil)
)

func TestSaveLoad_RoundTrip_WithComponentProvider(t *testing.T) {
	ecs := goke.New()
	module := &kinematicsModule{}
	ecs.RegSys(module)

	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&module.pos)
		factory.Create(2)
		for factory.Next() {
			positions := module.pos.Slice(&factory.Cursor)
			for i := range positions {
				positions[i] = Position{X: float32(i)}
			}
		}
		ids = append(ids, factory.IDs...)
	}})

	path := filepath.Join(t.TempDir(), "save.bin")
	ecs.Pause()
	if err := ecs.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	ecs2 := goke.New()
	module2 := &kinematicsModule{}
	// The caller only knows about its own module — not Position/Velocity by
	// name — yet Load still works via ComponentProvider/ProvidedComps.
	if err := ecs2.Load(path, goke.ProvidedComps(module2)...); err != nil {
		t.Fatalf("Load: %v", err)
	}

	var query *goke.Query
	ecs2.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		query = si.NewQueryBuilder(&module2.pos).Build()
	}})
	for i, id := range ids {
		p := seekComp(query, &module2.pos, id)
		if p == nil || p.X != float32(i) {
			t.Errorf("entity %d: Position = %+v", i, p)
		}
	}
}

func TestECS_Save_PanicsWithoutPause(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected Save to panic without a prior Pause()")
		}
	}()
	ecs := goke.New()
	_ = ecs.Save(filepath.Join(t.TempDir(), "save.bin"))
}

func TestECS_Tick_PanicsWhilePaused(t *testing.T) {
	ecs := goke.New()
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) { _ = ctx.Sync() })
	ecs.Pause()
	defer func() {
		if recover() == nil {
			t.Error("expected Tick to panic while paused")
		}
	}()
	ecs.Tick(time.Millisecond)
}

func TestECS_PauseResume_Idempotent(t *testing.T) {
	ecs := goke.New()
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) { _ = ctx.Sync() })

	ecs.Resume() // no-op when not paused
	if ecs.Paused() {
		t.Fatal("expected Paused() == false initially")
	}

	ecs.Pause()
	ecs.Pause() // idempotent
	if !ecs.Paused() {
		t.Fatal("expected Paused() == true after Pause")
	}

	ecs.Resume()
	ecs.Resume() // idempotent
	if ecs.Paused() {
		t.Fatal("expected Paused() == false after Resume")
	}
	ecs.Tick(time.Millisecond) // must not panic — not paused anymore
}

func TestECS_Load_PanicsIfNotFresh(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]() // any registration before Load

	defer func() {
		if recover() == nil {
			t.Error("expected Load to panic when called after other registration")
		}
	}()
	_ = ecs.Load(filepath.Join(t.TempDir(), "save.bin"), goke.RegisterFor[Position]())
}

type hasRawPointer struct{ V *int }

func TestECS_RegComp_PanicsOnNonEncodableType(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected RegComp to panic on a type with a raw pointer field")
		}
	}()
	ecs := goke.New()
	_ = ecs.RegComp[hasRawPointer]()
}

func TestECS_Reset_ClearsPaused(t *testing.T) {
	ecs := goke.New()
	ecs.Pause()
	ecs.Reset()
	if ecs.Paused() {
		t.Error("expected Reset to clear the paused state")
	}
}
