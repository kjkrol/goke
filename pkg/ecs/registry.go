package ecs

import (
	"reflect"
	"unsafe"
)

type Registry struct {
	entitiesRegistry   *entitiesRegistry
	componentsRegistry *componentsRegistry
	archetypeRegistry  *archetypeRegistry
}

func newRegistry() *Registry {
	componentsRegistry := newComponentsRegistry()
	return &Registry{
		entitiesRegistry:   newEntitiesRegistry(),
		componentsRegistry: componentsRegistry,
		archetypeRegistry:  newArchetypeRegistry(componentsRegistry),
	}
}

func (r *Registry) RemoveEntity(entity Entity) {
	mask, ok := r.entitiesRegistry.GetMask(entity)
	if !ok {
		return
	}

	arch := r.archetypeRegistry.Get(mask)
	if arch == nil {
		return
	}

	entityIndex, exists := arch.entityToIndex[entity]
	if !exists {
		return
	}
	arch.removeEntity(entityIndex)

	r.entitiesRegistry.destroy(entity)
}

func assign[T any](reg *Registry, entity Entity, component T) {
	componentType := reflect.TypeFor[T]()
	compID := reg.componentsRegistry.GetOrRegister(componentType)
	assignByID(reg, entity, compID, component)
}

func assignByID[T any](reg *Registry, entity Entity, compID ComponentID, component T) {
	oldMask, _ := reg.entitiesRegistry.GetMask(entity)

	newMask := oldMask.Set(compID)

	newArch := reg.archetypeRegistry.GetOrRegister(newMask)

	oldArch := reg.archetypeRegistry.Get(oldMask)

	reg.archetypeRegistry.MoveEntity(entity, oldArch, newArch, compID, unsafe.Pointer(&component))

	reg.entitiesRegistry.updateMask(entity, newMask)
}

func unassign[T any](reg *Registry, entity Entity) {
	componentType := reflect.TypeFor[T]()
	id, ok := reg.componentsRegistry.Get(componentType)
	if !ok {
		return
	}

	reg.unassignByID(entity, id)
}

func (r *Registry) unassignByID(entity Entity, compID ComponentID) {
	mask, maskExists := r.entitiesRegistry.GetMask(entity)
	if !maskExists || !mask.IsSet(compID) {
		return
	}

	newMask := mask.Clear(compID)
	oldArch := r.archetypeRegistry.Get(mask)
	newArch := r.archetypeRegistry.GetOrRegister(newMask)

	r.archetypeRegistry.MoveEntityOnly(entity, oldArch, newArch)

	r.entitiesRegistry.updateMask(entity, newMask)
}

func getComponent[T any](reg *Registry, entity Entity) *T {
	componentType := reflect.TypeFor[T]()
	componentID, ok := reg.componentsRegistry.Get(componentType)
	if !ok {
		return nil
	}

	mask, ok := reg.entitiesRegistry.GetMask(entity)
	if !ok {
		return nil
	}

	arch := reg.archetypeRegistry.Get(mask)
	if arch == nil {
		return nil
	}

	entityIndex, ok := arch.entityToIndex[entity]
	if !ok {
		return nil
	}

	col, ok := arch.columns[componentID]
	if !ok {
		return nil
	}

	return (*T)(col.GetElement(entityIndex))
}
