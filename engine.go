package goke

import (
	"time"

	"github.com/kjkrol/goke/internal/core"
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

// SetExecutionPlan defines the logic for each engine tick (how systems are orchestrated).
func (e *Engine) SetExecutionPlan(plan ExecutionPlan) {
	e.scheduler.SetExecutionPlan(plan)
}

// Tick updates the engine state by executing a single simulation step
// with the given delta time.
func (e *Engine) Tick(duration time.Duration) {
	e.scheduler.Tick(duration)
}
