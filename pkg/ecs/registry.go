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

	rec, ok := r.entitiesRegistry.GetRecord(entity)
	if !ok {
		return
	}
	if rec.arch == nil {
		return
	}

	rec.arch.removeEntity(rec.index)
	rec = nil
	r.entitiesRegistry.destroy(entity)
}

func assign[T any](reg *Registry, entity Entity, component T) {
	componentType := reflect.TypeFor[T]()
	compID := reg.componentsRegistry.GetOrRegister(componentType)
	assignByID(reg, entity, compID, component)
}

func assignByID[T any](reg *Registry, entity Entity, compID ComponentID, component T) error {
	rec, ok := reg.entitiesRegistry.GetRecord(entity)
	if !ok {
		return fmt.Errorf("Entity doesn't exist")
	}
	oldArch := rec.arch

	var oldMask ArchetypeMask
	if oldArch != nil {
		oldMask = oldArch.mask
	}

	newMask := oldMask.Set(compID)
	if oldMask == newMask {
		col := oldArch.columns[compID]
		col.setData(rec.index, unsafe.Pointer(&component))
		return nil
	}
	newArch := reg.archetypeRegistry.GetOrRegister(newMask)
	reg.archetypeRegistry.MoveEntity(reg, entity, oldArch, newArch, compID, unsafe.Pointer(&component))
	reg.entitiesRegistry.updateMask(entity, newMask)
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
	mask, maskExists := r.entitiesRegistry.GetMask(entity)
	if !maskExists || !mask.IsSet(compID) {
		return
	}

	newMask := mask.Clear(compID)
	oldArch := r.archetypeRegistry.Get(mask)
	newArch := r.archetypeRegistry.GetOrRegister(newMask)

	r.archetypeRegistry.MoveEntityOnly(r, entity, oldArch, newArch)

	r.entitiesRegistry.updateMask(entity, newMask)
}

func getComponent[T any](reg *Registry, entity Entity) *T {
	rec, ok := reg.entitiesRegistry.GetRecord(entity)
	if !ok {
		return nil
	}

	compType := reflect.TypeFor[T]()
	compID, _ := reg.componentsRegistry.Get(compType)

	col := rec.arch.columns[compID]
	return (*T)(col.GetElement(rec.index))
}
