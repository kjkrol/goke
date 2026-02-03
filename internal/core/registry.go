package core

import (
	"fmt"
	"reflect"
	"unsafe"
)

type Registry struct {
	EntityPool         *EntityGenerationalPool
	ComponentsRegistry *ComponentsRegistry
	ViewRegistry       *ViewRegistry
	ArchetypeRegistry  *ArchetypeRegistry
}

var _ ReadOnlyRegistry = (*Registry)(nil)

func NewRegistry(cfg RegistryConfig) *Registry {
	componentsRegistry := NewComponentsRegistry()
	viewRegistry := NewViewRegistry(cfg.ViewRegistryInitCap)
	archetypeRegistry := NewArchetypeRegistry(componentsRegistry, viewRegistry, cfg)
	return &Registry{
		EntityPool:         NewEntityGenerator(cfg.InitialEntityCap, cfg.FreeIndicesCap),
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

	r.ArchetypeRegistry.UnlinkEntity(entity)
	r.EntityPool.Release(entity)
	return true
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

func (r *Registry) ComponentGet(entity Entity, compID ComponentID) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entity) {
		return nil, fmt.Errorf("Invalid Entity")
	}

	link, ok := r.ArchetypeRegistry.EntityLinkStore.Get(entity)
	if !ok {
		return nil, fmt.Errorf("entity not found in registry")
	}

	col := link.Arch.Columns[compID]
	if col == nil {
		return nil, fmt.Errorf("component not found in archetype")
	}

	return col.GetElement(link.Row), nil
}
