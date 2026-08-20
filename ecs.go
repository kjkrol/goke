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
	paused    bool
	saving    bool
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
	adapter := &runnableAdapter{sys: system, wrapped: &CmdBuf{raw: raw}}
	ecs.scheduler.Register(adapter, raw)
	return adapter
}

// Setup runs each given system exactly once, in order: Init, then Update,
// then Sync — fully applying one system's effects (including deferred
// structural changes) before the next system's Init runs. Use it for
// one-time world seeding; a later system in the list can query and edit
// entities spawned by an earlier one. Callable only once per ECS lifetime
// (Reset starts a new one) — build everything you need in that single call.
func (ecs *ECS) Setup(systems ...System) {
	if ecs.setupDone {
		panic("goke: Setup called more than once — build everything you need (spawns, queries, editors) in a single Setup call")
	}
	ecs.setupDone = true
	for _, sys := range systems {
		sys.Init(&ecs.sysInit)
		raw := orch.NewCmdBuf()
		adapter := &runnableAdapter{sys: sys, wrapped: &CmdBuf{raw: raw}}
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
	if ecs.paused {
		panic("goke: Tick called while the ECS is paused — call Resume() first")
	}
	ecs.scheduler.Tick(duration)
}

// Pause stops Tick from running — a subsequent call panics until Resume.
// General-purpose (a host can use it as an ordinary game pause), and also
// the required precondition for Save: nothing may mutate the world while a
// snapshot is being written. Idempotent — calling it while already paused
// is a no-op.
func (ecs *ECS) Pause() { ecs.paused = true }

// Resume clears the paused state set by Pause, allowing Tick again.
// Idempotent — calling it while not paused is a no-op.
func (ecs *ECS) Resume() { ecs.paused = false }

// Paused reports whether the ECS is currently paused.
func (ecs *ECS) Paused() bool { return ecs.paused }

// Reset clears all entities, components, and system state, returning the ECS
// to its initial (post-New) condition. Registered component types are preserved.
// Also clears the paused state. Panics if called while a Save is in progress.
func (ecs *ECS) Reset() {
	if ecs.saving {
		panic("goke: Reset called while a Save is in progress")
	}
	ecs.scheduler.Reset()
	ecs.registry.Reset()
	ecs.setupDone = false
	ecs.paused = false
}
