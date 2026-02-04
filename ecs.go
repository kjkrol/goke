package goke

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/internal/core"
)

type (
	// Entity represents a 64-bit unique identifier for an object in the ECS world.
	Entity = core.Entity
	// ComponentID is a unique integer identifier for a specific component type.
	ComponentID = core.ComponentID
	// ComponentType contains metadata about a component type, such as its ID and memory size.
	ComponentType = core.ComponentInfo

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
	scheduler *core.SystemScheduler
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

	reg := core.NewRegistry(config)
	return &ECS{
		registry:  reg,
		scheduler: core.NewScheduler(reg),
	}
}

// ---- [E]CS Entity ----

// CreateEntity spawns a new entity within the registry and returns its identifier.
// The entity will have no components assigned initially.
func CreateEntity(ecs *ECS) Entity {
	return ecs.registry.CreateEntity()
}

// RemoveEntity destroys an entity and recycles its ID. All associated
// components are removed and memory is reclaimed. Returns true if the entity existed.
func RemoveEntity(ecs *ECS, entity Entity) bool {
	return ecs.registry.RemoveEntity(entity)
}

// SafeEnsureComponent attempts to return a direct pointer to the entity's component data.
// It functions as a zero-copy upsert: if the component exists, it returns a pointer
// to the existing data; otherwise, it allocates new memory.
// Returns an error if the entity does not exist or the allocation fails.
func SafeEnsureComponent[T any](ecs *ECS, entity Entity, compType ComponentType) (*T, error) {
	ptr, err := ecs.registry.AllocateByID(entity, compType)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// EnsureComponent returns a direct pointer to the entity's component data,
// providing the most efficient way to perform in-place upserts.
//
// Note: This function panics if the operation fails (e.g., if the entity is invalid).
// Use SafeEnsureComponent if you need to handle these errors gracefully.
func EnsureComponent[T any](ecs *ECS, entity Entity, compType ComponentType) *T {
	ptr, err := SafeEnsureComponent[T](ecs, entity, compType)
	if err != nil {
		panic(fmt.Sprintf("goke: failed to ensure component: %v", err))
	}
	return ptr
}

// GetComponent returns a typed pointer to an entity's component of type T.
func GetComponent[T any](ecs *ECS, entity Entity, compType ComponentType) (*T, error) {
	data, err := ecs.registry.ComponentGet(entity, compType.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}

// RemoveComponent removes a component from an entity using its ComponentInfo.
func RemoveComponent(ecs *ECS, entity Entity, compType ComponentType) error {
	return ecs.registry.UnassignByID(entity, compType)
}

// ---- E[C]S Component ----

// RegisterComponentType ensures a component of type T is registered in the ecs
// and returns its metadata.
func RegisterComponentType[T any](ecs *ECS) ComponentType {
	return core.EnsureComponentRegistered[T](ecs.registry.ComponentsRegistry)
}

// ---- EC[S] System ----

// RegisterSystem adds a stateful system to the ecs. The system's Init method
// will be called immediately.
func RegisterSystem(ecs *ECS, system System) {
	system.Init(ecs)
	ecs.scheduler.RegisterSystem(system)
}

// RegisterSystemFunc adds a stateless, function-based system to the ecs.
func RegisterSystemFunc(ecs *ECS, fn SystemFunc) System {
	wrapper := &functionalSystem{updateFn: fn}
	wrapper.Init(ecs)
	ecs.scheduler.RegisterSystem(wrapper)
	return wrapper
}

// ECS

// Plan defines the logic for each ecs tick (how systems are orchestrated).
func Plan(ecs *ECS, plan ExecutionPlan) {
	ecs.scheduler.SetExecutionPlan(plan)
}

// Tick updates the ecs state by executing a single simulation step
// with the given delta time.
func Tick(ecs *ECS, duration time.Duration) {
	ecs.scheduler.Tick(duration)
}
