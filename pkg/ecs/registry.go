package ecs

import (
	"fmt"
	"reflect"
	"unsafe"
)

const entityPoolInitCap = 1024
const viewRegistryInitCap = 1024
const archetypesInitCap = 64

type Registry struct {
	entityPool         *EntityGenerationalPool
	componentsRegistry *ComponentsRegistry
	viewRegistry       *ViewRegistry
	archetypeRegistry  *ArchetypeRegistry
}

var _ ReadOnlyRegistry = (*Registry)(nil)

func NewRegistry() *Registry {
	componentsRegistry := newComponentsRegistry()
	viewRegistry := NewViewRegistry()
	archetypeRegistry := newArchetypeRegistry(componentsRegistry, viewRegistry)
	return &Registry{
		entityPool:         NewEntityGenerator(entityPoolInitCap),
		componentsRegistry: componentsRegistry,
		viewRegistry:       viewRegistry,
		archetypeRegistry:  archetypeRegistry,
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

func (r *Registry) RegisterComponentType(componentType reflect.Type) ComponentInfo {
	return r.componentsRegistry.GetOrRegister(componentType)
}

func (r *Registry) GetComponent(entity Entity, compID ComponentID) (unsafe.Pointer, error) {
	if !r.entityPool.IsValid(entity) {
		return nil, fmt.Errorf("Invalid Entity")
	}

	link := r.archetypeRegistry.entityArchLinks[entity.Index()]
	col := link.arch.columns[compID]
	return col.GetElement(link.row), nil
}
