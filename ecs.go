package goke

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/orch"
	"github.com/kjkrol/goke/internal/reg"
)

var (
	_ orch.Lookup  = (*reg.Registry)(nil)
	_ orch.Mutator = (*reg.Registry)(nil)
)

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
	ecs.scheduler = orch.NewScheduler(&ecs.registry, &ecs.registry)
	return ecs
}

// ---- [E]CS Entity ----

// RemoveEntity destroys an entity and recycles its ID.
// All associated components are removed. Returns true if the entity existed.
func RemoveEntity(ecs *ECS, entity uid.UID64) bool {
	return ecs.registry.RemoveEntity(entity)
}

// ---- E[C]S Component ----

// UpsertComp returns a pointer to the entity's component, allocating it if absent.
// Panics if the entity is invalid or T does not match the registered type.
// Use SafeUpsertComp to handle errors explicitly.
func UpsertComp[T any](ecs *ECS, entity uid.UID64, compMeta CompMeta) *T {
	ptr, err := SafeUpsertComp[T](ecs, entity, compMeta)
	if err != nil {
		panic(fmt.Sprintf("goke: failed to upsert component: %v", err))
	}
	return ptr
}

// SafeUpsertComp returns a pointer to the entity's component, allocating it if absent.
// Returns ErrTypeMismatch if T does not match the registered type, or an error if the
// entity is invalid.
func SafeUpsertComp[T any](ecs *ECS, entity uid.UID64, compMeta CompMeta) (*T, error) {
	requestedType := reflect.TypeFor[T]()
	if requestedType != compMeta.Type {
		return nil, errTypeMismatch(compMeta.ID, compMeta.Type, requestedType)
	}
	ptr, err := ecs.registry.UpsertComp(entity, compMeta)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// SafeGetComp retrieves a typed pointer to an entity's component.
// Returns ErrTypeMismatch if T does not match the registered type, or an error if the
// entity or component is not found. Prefer for debugging or low-frequency access paths.
func SafeGetComp[T any](ecs *ECS, entity uid.UID64, compMeta CompMeta) (*T, error) {
	data, err := ecs.registry.GetComp(entity, compMeta.ID)
	requestedType := reflect.TypeFor[T]()
	if requestedType != compMeta.Type {
		return nil, errTypeMismatch(compMeta.ID, compMeta.Type, requestedType)
	}
	if err != nil {
		return nil, err
	}
	return (*T)(data), nil
}

// GetComp retrieves a typed pointer to an entity's component. Returns nil if the
// component is not found. Skips reflection checks — use only when T is known correct.
func GetComp[T any](ecs *ECS, entity uid.UID64, compMeta CompMeta) *T {
	data, err := ecs.registry.GetComp(entity, compMeta.ID)
	if err != nil {
		return nil
	}
	return (*T)(data)
}

// RemoveComp removes a component from an entity. Returns an error if the entity is invalid.
func RemoveComp(ecs *ECS, entity uid.UID64, compMeta CompMeta) error {
	return ecs.registry.RemoveComp(entity, compMeta)
}

// RegCompType registers the component type T with the ECS and returns its metadata.
// Call once at startup; subsequent calls for the same type return the cached metadata.
func RegCompType[T any](ecs *ECS) CompMeta {
	componentType := reflect.TypeFor[T]()
	return ecs.registry.RegCompType(componentType)
}

// ---- EC[S] System ----

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

// ---- ECS ----

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

// ---- EC[S] View ----

// NewView creates a View filtered by opts. Use Track[T]() to declare component
// data columns (accessible via Slice/At); Include[T]() for filter-only
// requirements; Exclude[T]() for exclusions.
func NewView(ecs *ECS, opts ...BlueprintOpt) *View {
	return ecs.registry.NewView(opts...)
}
