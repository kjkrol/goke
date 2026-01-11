package ecs

import (
	"time"
)

type (
	EngineAPI interface {
		CreateEntity() Entity
		RemoveEntity(entity Entity)

		UnassignByID(e Entity, id ComponentID)

		RegisterSystems(systems []System)
		UpdateSystems(duration time.Duration)
	}
	Engine struct {
		*Registry
		scheduler *scheduler
	}
)

func NewEngine() *Engine {
	reg := newRegistry()
	return &Engine{
		Registry:  reg,
		scheduler: newScheduler(reg),
	}
}

func (e *Engine) CreateEntity() Entity {
	return e.entitiesRegistry.create()
}

func (e *Engine) RegisterSystems(systems []System) {
	e.scheduler.registerSystems(systems)
}

func (e *Engine) UpdateSystems(duration time.Duration) {
	e.scheduler.updateSystems(duration)
}

func RegisterComponent[T any](e *Engine) ComponentID {
	return ensureComponentRegistered[T](e.Registry.componentsRegistry)
}

func Assign[T any](e *Engine, entity Entity, component T) {
	assign(e.Registry, entity, component)
}

func AssignByID[T any](e *Engine, entity Entity, id ComponentID, component T) {
	assignByID(e.Registry, entity, id, component)
}

func Unassign[T any](e *Engine, entity Entity) {
	unassign[T](e.Registry, entity)
}

func GetComponent[T any](e *Engine, entity Entity) *T {
	return getComponent[T](e.Registry, entity)
}
