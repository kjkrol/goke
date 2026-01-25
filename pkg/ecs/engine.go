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
		AssignByID(Entity, ComponentID, unsafe.Pointer) error
		UnassignByID(Entity, ComponentID) error
		GetComponent(Entity, ComponentID) (unsafe.Pointer, error)

		RegisterSystem(System)
		SetExecutionPlan(ExecutionPlan)
		UpdateSystems(time.Duration)
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

func (e *Engine) SetExecutionPlan(plan ExecutionPlan) {
	e.scheduler.SetExecutionPlan(plan)
}

func (e *Engine) UpdateSystems(duration time.Duration) {
	e.scheduler.UpdateSystems(duration)
}

func RegisterComponent[T any](eng *Engine) ComponentInfo {
	return ensureComponentRegistered[T](eng.componentsRegistry)
}

func Assign[T any](eng *Engine, entity Entity, component T) error {
	compInfo := ensureComponentRegistered[T](eng.componentsRegistry)
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compInfo.ID, data)
}

func AssignByID[T any](eng *Engine, entity Entity, compID ComponentID, component T) error {
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compID, data)
}

func Unassign[T any](eng *Engine, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	compInfo, ok := eng.componentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("Component doesn't exist.")
	}

	return eng.UnassignByID(entity, compInfo.ID)
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
