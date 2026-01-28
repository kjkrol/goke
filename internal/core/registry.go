package core

import (
	"fmt"
	"reflect"
	"unsafe"
)

const entityPoolInitCap = 1024
const viewRegistryInitCap = 1024
const archetypesInitCap = 64

type Registry struct {
	EntityPool         *EntityGenerationalPool
	ComponentsRegistry *ComponentsRegistry
	ViewRegistry       *ViewRegistry
	ArchetypeRegistry  *ArchetypeRegistry
}

var _ ReadOnlyRegistry = (*Registry)(nil)

func NewRegistry() *Registry {
	componentsRegistry := NewComponentsRegistry()
	viewRegistry := NewViewRegistry()
	archetypeRegistry := NewArchetypeRegistry(componentsRegistry, viewRegistry)
	return &Registry{
		EntityPool:         NewEntityGenerator(entityPoolInitCap),
		ComponentsRegistry: componentsRegistry,
		ViewRegistry:       viewRegistry,
		ArchetypeRegistry:  archetypeRegistry,
	}
}

func (r *Registry) CreateEntity() Entity {
	entity := r.EntityPool.Next()
	r.ArchetypeRegistry.AddEntity(entity)
	return entity
}

func (r *Registry) RemoveEntity(entity Entity) bool {
	if !r.EntityPool.IsValid(entity) {
		return false
	}

	r.ArchetypeRegistry.RemoveEntity(entity)
	r.EntityPool.Release(entity)
	return true
}

func (r *Registry) AssignByID(entity Entity, compInfo ComponentInfo, data unsafe.Pointer) error {
	if !r.EntityPool.IsValid(entity) {
		return fmt.Errorf("Invalid Entity")
	}

	r.ArchetypeRegistry.Assign(entity, compInfo, data)
	return nil
}

// AllocateByID ensures the entity is valid and performs the allocation in the archetype.
func (r *Registry) AllocateByID(entity Entity, compInfo ComponentInfo) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entity) {
		return nil, fmt.Errorf("invalid entity")
	}

	// Calling the new method we discussed for ArchetypeRegistry
	return r.ArchetypeRegistry.AllocateComponentMemory(entity, compInfo)
}

func (r *Registry) UnassignByID(entity Entity, compInfo ComponentInfo) error {
	if !r.EntityPool.IsValid(entity) {
		return fmt.Errorf("Invalid Entity")
	}

	r.ArchetypeRegistry.UnAssign(entity, compInfo)
	return nil
}

func (r *Registry) RegisterComponentType(componentType reflect.Type) ComponentInfo {
	return r.ComponentsRegistry.GetOrRegister(componentType)
}

func (r *Registry) GetComponent(entity Entity, compID ComponentID) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entity) {
		return nil, fmt.Errorf("Invalid Entity")
	}

	link := r.ArchetypeRegistry.EntityArchLinks[entity.Index()]
	col := link.Arch.Columns[compID]
	return col.GetElement(link.Row), nil
}
