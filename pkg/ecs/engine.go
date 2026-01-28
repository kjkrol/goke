package ecs

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/internal/core"
)

type Entity = core.Entity
type ComponentID = core.ComponentID
type ComponentInfo = core.ComponentInfo

type ExecutionPlan = core.ExecutionPlan
type ExecutionContext = core.ExecutionContext

type Engine struct {
	registry  *core.Registry
	scheduler *core.SystemScheduler
}

func NewEngine() *Engine {
	reg := core.NewRegistry()
	return &Engine{
		registry:  reg,
		scheduler: core.NewScheduler(reg),
	}
}

func (e *Engine) CreateEntity() Entity {
	return e.registry.CreateEntity()
}

func (e *Engine) RemoveEntity(entity Entity) bool {
	return e.registry.RemoveEntity(entity)
}

func (e *Engine) RegisterComponentType(componentType reflect.Type) ComponentInfo {
	return e.registry.RegisterComponentType(componentType)
}

func (e *Engine) Allocate(entity Entity, compInfo ComponentInfo) (unsafe.Pointer, error) {
	return e.registry.AllocateByID(entity, compInfo)
}

func (e *Engine) AssignByID(entity Entity, compInfo ComponentInfo, data unsafe.Pointer) error {
	return e.registry.AssignByID(entity, compInfo, data)
}

func (e *Engine) UnassignByID(entity Entity, compInfo ComponentInfo) error {
	return e.registry.UnassignByID(entity, compInfo)
}
func (e *Engine) GetComponent(entity Entity, compID ComponentID) (unsafe.Pointer, error) {
	return e.registry.GetComponent(entity, compID)
}

func (e *Engine) RegisterSystem(system System) {
	system.Init(e)
	e.scheduler.RegisterSystem(system)
}

func (e *Engine) RegisterSystemFunc(fn SystemFunc) System {
	wrapper := &functionalSystem{updateFn: fn}
	wrapper.Init(e)
	e.scheduler.RegisterSystem(wrapper)
	return wrapper
}

func (e *Engine) SetExecutionPlan(plan ExecutionPlan) {
	e.scheduler.SetExecutionPlan(plan)
}

func (e *Engine) Run(duration time.Duration) {
	e.scheduler.Tick(duration)
}

// generic helpers

func RegisterComponent[T any](eng *Engine) ComponentInfo {
	return core.EnsureComponentRegistered[T](eng.registry.ComponentsRegistry)
}

func Assign[T any](eng *Engine, entity Entity, component T) error {
	compInfo := core.EnsureComponentRegistered[T](eng.registry.ComponentsRegistry)
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compInfo, data)
}

func AssignByID[T any](eng *Engine, entity Entity, compInfo ComponentInfo, component T) error {
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compInfo, data)
}

func Unassign[T any](eng *Engine, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	compInfo, ok := eng.registry.ComponentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("Component doesn't exist.")
	}

	return eng.UnassignByID(entity, compInfo)
}

func AddComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compInfo := core.EnsureComponentRegistered[T](eng.registry.ComponentsRegistry)
	ptr, err := eng.Allocate(entity, compInfo)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

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
