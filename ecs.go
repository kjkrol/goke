package goke

import (
	"reflect"
	"time"

	"github.com/kjkrol/uid"

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
// Call once at startup; subsequent calls for the same type return the cached ID.
func RegComp[T any](ecs *ECS) CompID {
	compType := reflect.TypeFor[T]()
	return ecs.registry.RegComp(compType)
}

// RemoveEnt destroys an entity and recycles its ID.
// All associated components are removed. Returns true if the entity existed.
func (ecs *ECS) RemoveEnt(id uid.UID64) bool {
	return ecs.registry.Remove(id)
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
func (ecs *ECS) Tick(duration time.Duration) {
	ecs.scheduler.Tick(duration)
}

// Reset clears all entities, components, and system state, returning the ECS
// to its initial (post-New) condition. Registered component types are preserved.
func (ecs *ECS) Reset() {
	ecs.scheduler.Reset()
	ecs.registry.Reset()
	ecs.setupDone = false
}
