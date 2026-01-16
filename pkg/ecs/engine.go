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

		RegisterComponentType(reflect.Type) ComponentID
		AssignByID(Entity, ComponentID, unsafe.Pointer) error
		UnassignByID(Entity, ComponentID) error
		GetComponent(Entity, ComponentID) (unsafe.Pointer, error)

		RegisterSystems([]System)
		UpdateSystems(time.Duration)
	}
	Engine struct {
		*Registry
		scheduler *scheduler
	}
)

var _ EngineAPI = (*Engine)(nil)

func NewEngine() *Engine {
	reg := NewRegistry()
	return &Engine{
		Registry:  reg,
		scheduler: newScheduler(reg),
	}
}

func (e *Engine) RegisterSystems(systems []System) {
	e.scheduler.registerSystems(systems)
}

func (e *Engine) UpdateSystems(duration time.Duration) {
	e.scheduler.updateSystems(duration)
}

func RegisterComponent[T any](eng *Engine) ComponentID {
	return ensureComponentRegistered[T](eng.componentsRegistry)
}

func Assign[T any](eng *Engine, entity Entity, component T) error {
	compID := ensureComponentRegistered[T](eng.componentsRegistry)
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compID, data)
}

func AssignByID[T any](eng *Engine, entity Entity, compID ComponentID, component T) error {
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compID, data)
}

func Unassign[T any](eng *Engine, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	id, ok := eng.componentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("Component doesn't exist.")
	}

	return eng.UnassignByID(entity, id)
}

func UnassignByID[T any](eng *Engine, entity Entity, compID ComponentID) error {
	return eng.UnassignByID(entity, compID)
}

func GetComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compType := reflect.TypeFor[T]()
	compID, ok := eng.componentsRegistry.Get(compType)
	if !ok {
		return nil, fmt.Errorf("Component doesn't exist.")
	}
	data, err := eng.GetComponent(entity, compID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}
