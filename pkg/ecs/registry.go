package ecs

import (
	"fmt"
	"reflect"
	"unsafe"
)

const initialCapacity = 1024

type Registry struct {
	entityPool         *EntityGenerationalPool
	componentsRegistry *ComponentsRegistry
	archetypeRegistry  *ArchetypeRegistry
}

func NewRegistry() *Registry {
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

func (r *Registry) AssignByID(entity Entity, compID ComponentID, data unsafe.Pointer) error {
	if !r.entityPool.IsValid(entity) {
		return fmt.Errorf("Invalid Entity")
	}

	r.archetypeRegistry.Assign(entity, compID, data)
	return nil
}

func (r *Registry) UnassignByID(entity Entity, compID ComponentID) error {
	if !r.entityPool.IsValid(entity) {
		return fmt.Errorf("Invalid Entity")
	}

	r.archetypeRegistry.UnAssign(entity, compID)
	return nil
}

func (r *Registry) RegisterComponentType(componentType reflect.Type) ComponentID {
	return r.componentsRegistry.GetOrRegister(componentType)
}

func (r *Registry) GetComponent(entity Entity, compID ComponentID) (unsafe.Pointer, error) {
	if !r.entityPool.IsValid(entity) {
		return nil, fmt.Errorf("Invalid Entity")
	}

	backLink := r.archetypeRegistry.entityArchLinks[entity.Index()]
	col := backLink.arch.columns[compID]
	return col.GetElement(backLink.columnIndex), nil
}
