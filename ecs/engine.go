package ecs

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/ecs/internal/core"
)

type (
	// Entity represents a 64-bit unique identifier for an object in the ECS world.
	Entity = core.Entity
	// ComponentID is a unique integer identifier for a specific component type.
	ComponentID = core.ComponentID
	// ComponentInfo contains metadata about a component type, such as its ID and memory size.
	ComponentInfo = core.ComponentInfo

	// ExecutionPlan defines the order and concurrency of system updates.
	ExecutionPlan = core.ExecutionPlan
	// ExecutionContext provides methods to run systems (parallel or sync) within a plan.
	ExecutionContext = core.ExecutionContext

	EngineConfig = core.RegistryConfig
)

// Engine is the main entry point for the ECS. It acts as the coordinator
// that ties together data (entities and components) and logic (systems).
//
// Use the Engine to manage the lifecycle of entities, register component
// types, and define the execution flow of your application.
type Engine struct {
	registry  *core.Registry
	scheduler *core.SystemScheduler
}

// NewEngine creates and initializes a new ECS Engine instance.
// It accepts optional EngineOption functions to override the default EngineConfig,
// allowing for fine-tuned memory pre-allocation and performance optimization
// (e.g., adjusting archetype chunk sizes to minimize GC pressure).
func NewEngine(opts ...EngineOption) *Engine {
	config := EngineConfig{
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
	return &Engine{
		registry:  reg,
		scheduler: core.NewScheduler(reg),
	}
}

// CreateEntity spawns a new entity within the registry and returns its identifier.
// The entity will have no components assigned initially.
func (e *Engine) CreateEntity() Entity {
	return e.registry.CreateEntity()
}

// RemoveEntity destroys an entity and recycles its ID. All associated
// components are removed and memory is reclaimed. Returns true if the entity existed.
func (e *Engine) RemoveEntity(entity Entity) bool {
	return e.registry.RemoveEntity(entity)
}

// RegisterComponentType manually registers a reflect.Type as a component.
// Most users should use the generic RegisterComponent function instead.
func (e *Engine) RegisterComponentType(componentType reflect.Type) ComponentInfo {
	return e.registry.RegisterComponentType(componentType)
}

// RemoveComponentByID removes a component from an entity using its ComponentInfo.
func (e *Engine) RemoveComponentByID(entity Entity, compInfo ComponentInfo) error {
	return e.registry.UnassignByID(entity, compInfo)
}

// GetComponent returns a raw pointer to an entity's component.
// Returns an error if the entity does not have the component.
func (e *Engine) GetComponent(entity Entity, compID ComponentID) (unsafe.Pointer, error) {
	return e.registry.GetComponent(entity, compID)
}

// RegisterSystem adds a stateful system to the engine. The system's Init method
// will be called immediately.
func (e *Engine) RegisterSystem(system System) {
	system.Init(e)
	e.scheduler.RegisterSystem(system)
}

// RegisterSystemFunc adds a stateless, function-based system to the engine.
func (e *Engine) RegisterSystemFunc(fn SystemFunc) System {
	wrapper := &functionalSystem{updateFn: fn}
	wrapper.Init(e)
	e.scheduler.RegisterSystem(wrapper)
	return wrapper
}

// SetExecutionPlan defines the logic for each engine tick (how systems are orchestrated).
func (e *Engine) SetExecutionPlan(plan ExecutionPlan) {
	e.scheduler.SetExecutionPlan(plan)
}

// Tick updates the engine state by executing a single simulation step
// with the given delta time.
func (e *Engine) Tick(duration time.Duration) {
	e.scheduler.Tick(duration)
}

// --- Generic Helpers ---

// RegisterComponent ensures a component of type T is registered in the engine
// and returns its metadata.
func RegisterComponent[T any](eng *Engine) ComponentInfo {
	return core.EnsureComponentRegistered[T](eng.registry.ComponentsRegistry)
}

// RemoveComponent removes a component of type T from an entity.
func RemoveComponent[T any](eng *Engine, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	compInfo, ok := eng.registry.ComponentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("Component doesn't exist.")
	}

	return eng.registry.UnassignByID(entity, compInfo)
}

// AllocateComponent reserves memory for a component of type T and returns a direct pointer.
// It is the most efficient way to add data, enabling in-place modification and
// bypassing temporary copies or potential heap escapes.
func AllocateComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compInfo := core.EnsureComponentRegistered[T](eng.registry.ComponentsRegistry)
	ptr, err := eng.registry.AllocateByID(entity, compInfo)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// AllocateComponentByInfo assigns a component to an entity using pre-cached ComponentInfo.
// This is a high-performance alternative to AllocateComponent, as it skips
// registry lookups and type reflection. It returns a typed pointer for
// direct in-place initialization.
func AllocateComponentByInfo[T any](eng *Engine, entity Entity, compInfo ComponentInfo) (*T, error) {
	ptr, err := eng.registry.AllocateByID(entity, compInfo)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// AllocateComponent reserves memory for a component of type T and returns a direct pointer.
// It is the most efficient way to add data, enabling in-place modification and
// bypassing temporary copies or potential heap escapes.
func SetComponent[T any](eng *Engine, entity Entity, value T) error {
	if ptr, err := AllocateComponent[T](eng, entity); err == nil {
		*ptr = value
		return nil
	} else {
		return err
	}
}

// GetComponent returns a typed pointer to an entity's component of type T.
func GetComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compType := reflect.TypeFor[T]()
	compInfo, ok := eng.registry.ComponentsRegistry.Get(compType)
	if !ok {
		return nil, fmt.Errorf("Component doesn't exist.")
	}
	data, err := eng.registry.GetComponent(entity, compInfo.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}

// GetComponent returns a typed pointer to an entity's component of type T.
func GetComponentByType[T any](eng *Engine, entity Entity, compInfo ComponentInfo) (*T, error) {
	data, err := eng.registry.GetComponent(entity, compInfo.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}
