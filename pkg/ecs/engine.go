package ecs

import (
	"time"
)

type (
	Entity uint32

	ComponentID int

	System interface {
		Init(*Registry)
		Update(*Registry, time.Duration)
	}

	View struct {
		mask Bitmask
	}

	Engine struct {
		registry  *Registry
		scheduler *scheduler
	}
)

func NewEngine() *Engine {
	reg := newRegistry()
	return &Engine{
		registry:  reg,
		scheduler: newScheduler(reg),
	}
}

func (e *Engine) CreateEntity() Entity {
	return e.registry.createEntity()
}

func (e *Engine) RemoveEntity(entity Entity) {
	e.registry.removeEntity(entity)
}

func (e *Engine) RegisterSystems(systems []System) {
	e.scheduler.registerSystems(systems)
}

func (e *Engine) UpdateSystems(duration time.Duration) {
	e.scheduler.updateSystems(duration)
}

func RegisterComponent[T any](e *Engine) ComponentID {
	return registerComponent[T](e.registry)
}

func Assign[T any](e *Engine, entity Entity, component T) {
	assign(e.registry, entity, component)
}

func AssignByID[T any](e *Engine, entity Entity, id ComponentID, component T) {
	assignByID(e.registry, entity, id, component)
}

func Unassign[T any](e *Engine, entity Entity) {
	unassign[T](e.registry, entity)
}

func UnassignByID[T any](e *Engine, entity Entity, id ComponentID) {
	unassignByID[T](e.registry, entity, id)
}

func Get[T any](e *Engine, entity Entity) *T { return get[T](e.registry, entity) }
