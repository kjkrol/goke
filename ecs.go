package goke

import (
	"fmt"
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

// CreateFactory resolves or creates the archetype from Track opts and returns
// a reusable Factory ready for repeated Create/Next cycles.
func CreateFactory(ecs *ECS, opts ...Opt) *Factory {
	return ecs.registry.CreateFactory(opts...)
}

// CreateView creates a View filtered by opts. Use Track[T]() to declare component
// data columns (accessible via Slice/At); Include[T]() for filter-only
// requirements; Exclude[T]() for exclusions.
func CreateView(ecs *ECS, opts ...Opt) *View {
	return ecs.registry.AddView(opts...)
}

// CreateLookup creates a Lookup for single-entity component access.
// Use Track[T]() opts to declare which components are accessible via col.At.
// Call Seek on each access; the cursor is positioned at the entity's storage slot.
func CreateLookup(ecs *ECS, opts ...Opt) *Lookup {
	return ecs.registry.CreateLookup(opts...)
}

// UpsertComp returns a pointer to the entity's component, allocating it if absent.
// Panics if the entity is invalid.
func UpsertComp[T any](ecs *ECS, uid uid.UID64, compID CompID) *T {
	ptr, err := ecs.registry.UpsertComp(uid, compID)
	if err != nil {
		panic(fmt.Sprintf("goke: failed to upsert component: %v", err))
	}
	return (*T)(ptr)
}

// RemoveComp removes a component from an entity id. Returns an error if the entity is invalid.
func RemoveComp(ecs *ECS, uid uid.UID64, compID CompID) error {
	return ecs.registry.RemoveComp(uid, compID)
}

// RemoveEntity destroys an entity and recycles its ID.
// All associated components are removed. Returns true if the entity existed.
func RemoveEnt(ecs *ECS, uid uid.UID64) bool {
	return ecs.registry.Remove(uid)
}

// RegSys registers a stateful system. The system's Init method is called immediately.
func RegSys(ecs *ECS, system System) {
	system.Init(ecs)
	ecs.scheduler.Register(system)
}

// RegSysFn registers a stateless function as a system and returns the created System.
func RegSysFn(ecs *ECS, fn SystemFn) System {
	wrapper := &functionalSystem{updateFn: fn}
	wrapper.Init(ecs)
	ecs.scheduler.Register(wrapper)
	return wrapper
}

// SetPlan sets the execution plan that controls how systems run each tick.
// Call before the first Tick; replaces any previously set plan.
func SetPlan(ecs *ECS, plan Plan) {
	ecs.scheduler.SetPlan(plan)
}

// Tick advances the simulation by one step with the given delta time.
func Tick(ecs *ECS, duration time.Duration) {
	ecs.scheduler.Tick(duration)
}

// Reset clears all entities, components, and system state, returning the ECS
// to its initial (post-New) condition. Registered component types are preserved.
func Reset(ecs *ECS) {
	ecs.scheduler.Reset()
	ecs.registry.Reset()
}
