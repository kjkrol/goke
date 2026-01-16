package ecs

import (
	"fmt"
	"reflect"
	"unsafe"
)

const initialCapacity = 1024

type Registry struct {
	entityPool         *EntityGenerationalPool
	componentsRegistry *componentsRegistry
	archetypeRegistry  *archetypeRegistry
}

type EntityArchLink struct {
	arch        *Archetype
	columnIndex int
}

func newRegistry() *Registry {
	componentsRegistry := newComponentsRegistry()
	return &Registry{
		entityPool:         NewEntityGenerator(initialCapacity),
		componentsRegistry: componentsRegistry,
		archetypeRegistry:  newArchetypeRegistry(componentsRegistry),
	}
}

func (r *Registry) CreateEntity() Entity {
	entity := r.entityPool.Next()
	r.archetypeRegistry.AddEntity(entity)
	return entity
}

func (r *Registry) RemoveEntity(entity Entity) bool {
	if !r.entityPool.IsValid(entity) {
		return false
	}
	r.archetypeRegistry.RemoveEntity(entity)
	r.entityPool.Release(entity)
	return true
}

func assign[T any](reg *Registry, entity Entity, component T) error {
	componentType := reflect.TypeFor[T]()
	compID := reg.componentsRegistry.GetOrRegister(componentType)
	return assignByID(reg, entity, compID, component)
}

func assignByID[T any](reg *Registry, entity Entity, compID ComponentID, component T) error {
	if !reg.entityPool.IsValid(entity) {
		return fmt.Errorf("Invalid Entity")
	}
	data := unsafe.Pointer(&component)
	reg.archetypeRegistry.Assign(entity, compID, data)
	return nil
}

func unassign[T any](reg *Registry, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	id, ok := reg.componentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("Component doesn't exist.")
	}

	return reg.unassignByID(entity, id)
}

func (r *Registry) unassignByID(entity Entity, compID ComponentID) error {
	if !r.entityPool.IsValid(entity) {
		return fmt.Errorf("Invalid Entity")
	}

	r.archetypeRegistry.UnAssign(entity, compID)
	return nil
}

func getComponent[T any](reg *Registry, entity Entity) (*T, error) {
	compType := reflect.TypeFor[T]()
	compID, ok := reg.componentsRegistry.Get(compType)
	if !ok {
		return nil, fmt.Errorf("Component doesn't exist.")
	}

	if !reg.entityPool.IsValid(entity) {
		return nil, fmt.Errorf("Invalid Entity")
	}

	backLink := reg.archetypeRegistry.entityArchLinks[entity.Index()]

	col := backLink.arch.columns[compID]
	return (*T)(col.GetElement(backLink.columnIndex)), nil
}
