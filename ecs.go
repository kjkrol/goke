package goke

import (
	"reflect"
	"time"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/orch"
	"github.com/kjkrol/goke/internal/reg"
)

var _ orch.Mutator = (*reg.Registry)(nil)

// ECS is the central coordinator of the entity-component-system world.
// It manages entity lifecycles, component storage, and system execution.
type ECS struct {
	registry  reg.Registry
	scheduler orch.Scheduler
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
	return ecs
}

// RegComp registers the component type T with the ECS and returns its ID.
// Call once at startup; subsequent calls for the same type return the cached ID.
func RegComp[T any](ecs *ECS) CompID {
	compType := reflect.TypeFor[T]()
	return ecs.registry.RegComp(compType)
}

// NewFactory resolves or creates the archetype from the given components and
// returns a reusable Factory ready for repeated Create/Next cycles. Each
// component behaves like Add[T] — pass &comp directly.
func (ecs *ECS) NewFactory(comps ...Addable) *Factory {
	opts := make([]EditOpt, len(comps))
	for i, c := range comps {
		opts[i] = c.asAdd()
	}
	return ecs.registry.CreateFactory(opts...)
}

// RemoveEnt destroys an entity and recycles its ID.
// All associated components are removed. Returns true if the entity existed.
func (ecs *ECS) RemoveEnt(id uid.UID64) bool {
	return ecs.registry.Remove(id)
}

// RegSys registers a stateful system. The system's Init method is called immediately.
func (ecs *ECS) RegSys(system System) {
	system.Init(ecs)
	ecs.scheduler.Register(system)
}

// RegSysFn registers a stateless function as a system and returns the created System.
func (ecs *ECS) RegSysFn(fn SystemFn) System {
	wrapper := &functionalSystem{updateFn: fn}
	wrapper.Init(ecs)
	ecs.scheduler.Register(wrapper)
	return wrapper
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
}
