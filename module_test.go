package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
)

type ModuleTag struct{ N int }

// fakeModule is a minimal goke.Module: one Setup-phase spawn, one per-tick
// system, and a component it owns.
type fakeModule struct {
	tag      goke.Comp[ModuleTag]
	sys      goke.System
	handle   goke.Runnable
	seeded   bool
	ticked   int
	spawnErr error
}

func (m *fakeModule) RegSystems(ecs *goke.ECS) {
	m.sys = goke.SystemFn{OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) { m.ticked++ }}
	m.handle = ecs.RegSys(m.sys)
}

func (m *fakeModule) RunPlan(ctx goke.RunCtx, d time.Duration) {
	ctx.Run(m.handle, d)
}

func (m *fakeModule) SetupSystems() []goke.System {
	return []goke.System{goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&m.tag)
		factory.Create(1)
		m.seeded = factory.Next()
	}}}
}

func (m *fakeModule) LoadComps() []goke.CompToken {
	return goke.LoadComps(&m.tag)
}

var _ goke.Module = (*fakeModule)(nil)

func TestModule_RegModuleSetupAndRunPlan(t *testing.T) {
	ecs := goke.New()
	m := &fakeModule{}

	ecs.RegModule(m)
	ecs.Setup(m.SetupSystems()...)

	if !m.seeded {
		t.Fatal("expected SetupSystems to spawn an entity")
	}

	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) { m.RunPlan(ctx, d) })
	ecs.Tick(time.Millisecond)
	ecs.Tick(time.Millisecond)

	if m.ticked != 2 {
		t.Errorf("expected RunPlan's system to have ticked 2 times, got %d", m.ticked)
	}
}
