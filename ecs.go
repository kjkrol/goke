package goke

import (
	"reflect"
	"time"

	"github.com/kjkrol/goke/v3/internal/orch"
	"github.com/kjkrol/goke/v3/internal/reg"
)

var _ orch.Mutator = (*reg.Registry)(nil)

// ECS is the central coordinator of the entity-component-system world.
// It manages entity lifecycles, component storage, and system execution.
type ECS struct {
	registry  reg.Registry
	scheduler orch.Scheduler
	sysInit   SysInit
	setupDone bool
}

// New creates a new ECS instance. Use ECSOption functions to tune memory
// pre-allocation for your expected entity count and component variety.
func New(opts ...ECSOption) *ECS {
	config := reg.DefaultConfig()

	for _, opt := range opts {
		opt(&config)
	}

	ecs := &ECS{}
	ecs.registry.Init(config)
	ecs.scheduler = orch.NewScheduler(&ecs.registry)
	ecs.sysInit = SysInit{ecs: ecs}
	return ecs
}

// RegComp registers the component type T with the ECS and returns its ID.
// Call once at startup; subsequent calls for the same type return the
// cached ID. Panics if T isn't encodable — see [Comp] for the exact rule.
func (ecs *ECS) RegComp[T any]() CompID {
	compType := reflect.TypeFor[T]()
	return ecs.registry.RegComp(compType)
}

// RegSys registers a system. The system's Init method is called
// immediately. Returns a Runnable handle — pass it to RunCtx.Run/RunParallel
// inside a Plan.
func (ecs *ECS) RegSys(system System) Runnable {
	system.Init(&ecs.sysInit)
	raw := orch.NewCmdBuf()
	wrapped := &CmdBuf{raw: raw}
	fn := orch.RunnableFunc(func(_ *orch.CmdBuf, d time.Duration) {
		system.Update(wrapped, d)
	})
	adapter := &fn
	ecs.scheduler.Register(adapter, raw)
	return adapter
}

// RegModule registers a Module's systems — equivalent to calling RegSys on
// each of them, except the module keeps the resulting handles to itself
// and replays them, in the required order, from RunPlan.
func (ecs *ECS) RegModule(m Module) {
	m.RegSystems(ecs)
}

// Setup runs each given system once — Init, then Update, then Sync — fully
// applying its effects before the next system's Init runs, so later systems
// see what earlier ones spawned. Callable once per ECS lifetime (Reset
// starts a new one).
func (ecs *ECS) Setup(systems ...System) {
	if ecs.setupDone {
		panic("goke: Setup called more than once — build everything you need (spawns, queries, editors) in a single Setup call")
	}
	ecs.setupDone = true
	for _, sys := range systems {
		sys.Init(&ecs.sysInit)
		raw := orch.NewCmdBuf()
		wrapped := &CmdBuf{raw: raw}
		fn := orch.RunnableFunc(func(_ *orch.CmdBuf, d time.Duration) {
			sys.Update(wrapped, d)
		})
		adapter := &fn
		sched := orch.NewScheduler(&ecs.registry)
		sched.Register(adapter, raw)
		sched.SetPlan(func(ctx orch.RunCtx, d time.Duration) {
			ctx.Run(adapter, d)
			_ = ctx.Sync()
		})
		sched.Tick(0)
	}
}

// SetPlan sets the execution plan that controls how systems run each tick.
// Call before the first Tick; replaces any previously set plan.
func (ecs *ECS) SetPlan(plan Plan) {
	ecs.scheduler.SetPlan(plan)
}

// Tick advances the simulation by one step with the given delta time.
// Panics if the ECS is paused (see [ECS.Pause]) — call [ECS.Resume] first.
func (ecs *ECS) Tick(duration time.Duration) {
	if ecs.registry.Paused() {
		panic("goke: Tick called while the ECS is paused — call Resume() first")
	}
	ecs.scheduler.Tick(duration)
}

// Pause stops Tick from running (panics until Resume) — also required
// before Save. Idempotent.
func (ecs *ECS) Pause() { ecs.registry.Pause() }

// Resume clears the paused state set by Pause, allowing Tick again.
// Idempotent — calling it while not paused is a no-op.
func (ecs *ECS) Resume() { ecs.registry.Resume() }

// Paused reports whether the ECS is currently paused.
func (ecs *ECS) Paused() bool { return ecs.registry.Paused() }

// Reset clears all entities, components, and system state, returning the ECS
// to its initial (post-New) condition. Registered component types are preserved.
// Also clears the paused state. Panics if called while a Save is in progress.
func (ecs *ECS) Reset() {
	ecs.scheduler.Reset()
	ecs.registry.Reset()
	ecs.setupDone = false
}

// Save writes a full snapshot of the world to path — every entity and
// component, with original IDs preserved. Requires a prior [ECS.Pause]
// (panics otherwise).
func (ecs *ECS) Save(path string) error { return ecs.registry.Save(path) }

// Load reads a snapshot written by Save into ecs — components, archetypes,
// and entities, with original IDs. Must run before Setup or any other
// registration (panics otherwise); matches comps by name, any order — see
// [LoadComp], [CompProvider].
func (ecs *ECS) Load(path string, comps ...CompToken) error {
	return ecs.registry.Load(path, comps)
}
