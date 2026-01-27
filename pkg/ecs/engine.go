package ecs

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

type (
	EngineAPI interface {
		CreateEntity() Entity
		RemoveEntity(entity Entity) bool

		RegisterComponentType(reflect.Type) ComponentInfo
		AssignByID(Entity, ComponentInfo, unsafe.Pointer) error
		AllocateComponentMemoryByID(Entity, ComponentInfo) (unsafe.Pointer, error)
		UnassignByID(Entity, ComponentInfo) error
		GetComponent(Entity, ComponentID) (unsafe.Pointer, error)

		RegisterSystem(System)
		RegisterSystemFunc(SystemFunc) System
		SetExecutionPlan(ExecutionPlan)
		Run(time.Duration)
	}
	Engine struct {
		*Registry
		scheduler *SystemScheduler
	}
)

var _ EngineAPI = (*Engine)(nil)

func NewEngine() *Engine {
	reg := NewRegistry()
	return &Engine{
		Registry:  reg,
		scheduler: NewScheduler(reg),
	}
}

func (e *Engine) RegisterSystem(system System) {
	e.scheduler.RegisterSystem(system)
}

func (e *Engine) RegisterSystemFunc(fn SystemFunc) System {
	wrapper := &functionalSystem{updateFn: fn}
	e.scheduler.RegisterSystem(wrapper)
	return wrapper
}

func (e *Engine) SetExecutionPlan(plan ExecutionPlan) {
	e.scheduler.SetExecutionPlan(plan)
}

func (e *Engine) Run(duration time.Duration) {
	e.scheduler.Tick(duration)
}

func RegisterComponent[T any](eng *Engine) ComponentInfo {
	return ensureComponentRegistered[T](eng.componentsRegistry)
}

func Assign[T any](eng *Engine, entity Entity, component T) error {
	compInfo := ensureComponentRegistered[T](eng.componentsRegistry)
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compInfo, data)
}

func AssignByID[T any](eng *Engine, entity Entity, compInfo ComponentInfo, component T) error {
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compInfo, data)
}

func Unassign[T any](eng *Engine, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	compInfo, ok := eng.componentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("Component doesn't exist.")
	}

	return eng.UnassignByID(entity, compInfo)
}

func AddComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compInfo := ensureComponentRegistered[T](eng.componentsRegistry)
	ptr, err := eng.AllocateComponentMemoryByID(entity, compInfo)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

func GetComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compType := reflect.TypeFor[T]()
	compInfo, ok := eng.componentsRegistry.Get(compType)
	if !ok {
		return nil, fmt.Errorf("Component doesn't exist.")
	}
	data, err := eng.GetComponent(entity, compInfo.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}
