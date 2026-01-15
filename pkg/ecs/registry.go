package ecs

import (
	"fmt"
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
	backLink, ok := r.entitiesRegistry.GetBackLink(entity)
	if !ok {
		return
	}
	if backLink.arch == nil {
		return
	}

	backLink.arch.removeEntity(backLink.index)
	backLink = nil
	r.entitiesRegistry.destroy(entity)
}

func assign[T any](reg *Registry, entity Entity, component T) {
	componentType := reflect.TypeFor[T]()
	compID := reg.componentsRegistry.GetOrRegister(componentType)
	assignByID(reg, entity, compID, component)
}

func assignByID[T any](reg *Registry, entity Entity, compID ComponentID, component T) error {
	backLink, ok := reg.entitiesRegistry.GetBackLink(entity)
	if !ok {
		return fmt.Errorf("Entity doesn't exist")
	}
	oldArch := backLink.arch

	var oldMask ArchetypeMask
	if oldArch != nil {
		oldMask = oldArch.mask
	}

	newMask := oldMask.Set(compID)
	if oldMask == newMask {
		col := oldArch.columns[compID]
		col.setData(backLink.index, unsafe.Pointer(&component))
		return nil
	}
	newArch := reg.archetypeRegistry.GetOrRegister(newMask)
	reg.archetypeRegistry.MoveEntity(reg, entity, oldArch, newArch, compID, unsafe.Pointer(&component))
	return nil
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
	backLink, ok := r.entitiesRegistry.GetBackLink(entity)
	if !ok {
		return
	}

	oldArch := backLink.arch
	mask := oldArch.mask

	newMask := mask.Clear(compID)
	newArch := r.archetypeRegistry.GetOrRegister(newMask)

	r.archetypeRegistry.MoveEntityOnly(r, entity, oldArch, newArch)
}

func getComponent[T any](reg *Registry, entity Entity) *T {
	backLink, ok := reg.entitiesRegistry.GetBackLink(entity)
	if !ok {
		return nil
	}

	compType := reflect.TypeFor[T]()
	compID, _ := reg.componentsRegistry.Get(compType)

	col := backLink.arch.columns[compID]
	return (*T)(col.GetElement(backLink.index))
}
