package goke

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kjkrol/goke/internal/core"
)

type (
	// Entity represents a 64-bit unique identifier for an object in the ECS world.
	Entity = core.Entity
	// ComponentID is a unique integer identifier for a specific component type.
	ComponentID = core.ComponentID
	// ComponentDesc is a component descriptor; it contains metadata about a component type, such as its ID and memory size.
	ComponentDesc = core.ComponentInfo

	ECSConfig = core.RegistryConfig
	// ExecutionContext provides methods to run systems (parallel or sync) within a plan.
	ExecutionContext = core.ExecutionContext
	// ExecutionPlan defines the order and concurrency of system updates.
	ExecutionPlan = core.ExecutionPlan
)

// ECS is the main entry point for the ECS. It acts as the coordinator
// that ties together data (entities and components) and logic (systems).
//
// Use the ECS to manage the lifecycle of entities, register component
// types, and define the execution flow of your application.
type ECS struct {
	registry  *core.Registry
	scheduler core.SystemScheduler
}

// New creates and initializes a new ECS instance.
// It accepts optional ECSOption functions to override the default ECSConfig,
// allowing for fine-tuned memory pre-allocation and performance optimization
// (e.g., adjusting archetype chunk sizes to minimize GC pressure).
func New(opts ...ECSOption) *ECS {
	config := ECSConfig{
		InitialEntityCap:            1024,
		DefaultArchetypeChunkSize:   128,
		InitialArchetypeRegistryCap: 64,
		FreeIndicesCap:              1024,
		ViewRegistryInitCap:         32,
	}

	for _, opt := range opts {
		opt(&config)
	}

	ecs := ECS{
		registry: core.NewRegistry(config),
	}
	ecs.scheduler = core.NewScheduler(ecs.registry)
	return &ecs
}

// ---- [E]CS Entity ----

// RemoveEntity destroys an entity and recycles its ID. All associated
// components are removed and memory is reclaimed. Returns true if the entity existed.
func RemoveEntity(ecs *ECS, entity Entity) bool {
	return ecs.registry.RemoveEntity(entity)
}

// EnsureComponent returns a direct pointer to the entity's component data,
// providing the most efficient way to perform in-place upserts.
//
// Note: This function panics if the operation fails (e.g., if the entity is invalid).
// Use SafeEnsureComponent if you need to handle these errors gracefully.
func EnsureComponent[T any](ecs *ECS, entity Entity, compDesc ComponentDesc) *T {
	ptr, err := SafeEnsureComponent[T](ecs, entity, compDesc)
	if err != nil {
		panic(fmt.Sprintf("goke: failed to ensure component: %v", err))
	}
	return ptr
}

// SafeEnsureComponent attempts to return a direct pointer to the entity's component data.
// It functions as a zero-copy upsert: if the component exists, it returns a pointer
// to the existing data; otherwise, it allocates new memory.
// Returns an error if the entity does not exist or the allocation fails.
func SafeEnsureComponent[T any](ecs *ECS, entity Entity, compDesc ComponentDesc) (*T, error) {
	requestedType := reflect.TypeFor[T]()
	if requestedType != compDesc.Type {
		return nil, fmt.Errorf("type mismatch: component ID %d is registered as %v, but requested as %v",
			compDesc.ID, compDesc.Type, requestedType)
	}
	ptr, err := ecs.registry.AllocateByID(entity, compDesc)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// GetComponent returns a typed pointer to an entity's component of type T.
func GetComponent[T any](ecs *ECS, entity Entity, compDesc ComponentDesc) (*T, error) {
	data, err := ecs.registry.ComponentGet(entity, compDesc.ID)
	requestedType := reflect.TypeFor[T]()
	if requestedType != compDesc.Type {
		return nil, fmt.Errorf("type mismatch: component ID %d is registered as %v, but requested as %v",
			compDesc.ID, compDesc.Type, requestedType)
	}
	if err != nil {
		return nil, err
	}
	return (*T)(data), nil
}

// RemoveComponent removes a component from an entity using its ComponentInfo.
func RemoveComponent(ecs *ECS, entity Entity, compDesc ComponentDesc) error {
	return ecs.registry.UnassignByID(entity, compDesc)
}

// ---- E[C]S Component ----

// RegisterComponent ensures a component of type T is registered in the ECS
// and returns its metadata.
func RegisterComponent[T any](ecs *ECS) ComponentDesc {
	return core.EnsureComponentRegistered[T](&ecs.registry.ComponentsRegistry)
}

// ---- EC[S] System ----

// RegisterSystem adds a stateful system to the ECS. The system's Init method
// will be called immediately.
func RegisterSystem(ecs *ECS, system System) {
	system.Init(ecs)
	ecs.scheduler.RegisterSystem(system)
}

// RegisterSystemFunc adds a stateless, function-based system to the ECS.
func RegisterSystemFunc(ecs *ECS, fn SystemFunc) System {
	wrapper := &functionalSystem{updateFn: fn}
	wrapper.Init(ecs)
	ecs.scheduler.RegisterSystem(wrapper)
	return wrapper
}

// ---- ECS ----

// Plan defines the logic for each ECS tick (how systems are orchestrated).
func Plan(ecs *ECS, plan ExecutionPlan) {
	ecs.scheduler.SetExecutionPlan(plan)
}

// Tick updates the ecs state by executing a single simulation step
// with the given delta time.
func Tick(ecs *ECS, duration time.Duration) {
	ecs.scheduler.Tick(duration)
}
