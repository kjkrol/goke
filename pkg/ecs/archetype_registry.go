package ecs

import (
	"reflect"
	"unsafe"
)

const initArchetypesCapacity = 64

type archetypeRegistry struct {
	archetypeMap       map[ArchetypeMask]*archetype
	archetypes         []*archetype
	componentsRegistry *componentsRegistry
}

func newArchetypeRegistry(componentsRegistry *componentsRegistry) *archetypeRegistry {
	return &archetypeRegistry{
		archetypeMap:       make(map[ArchetypeMask]*archetype),
		archetypes:         make([]*archetype, 0, initArchetypesCapacity),
		componentsRegistry: componentsRegistry,
	}
}

func (r *archetypeRegistry) Get(mask ArchetypeMask) *archetype {
	if arch, ok := r.archetypeMap[mask]; ok {
		return arch
	}
	return nil
}

func (r *archetypeRegistry) All() []*archetype {
	return r.archetypes
}

func (r *archetypeRegistry) GetOrRegister(mask ArchetypeMask) *archetype {
	if arch, ok := r.archetypeMap[mask]; ok {
		return arch
	}

	arch := newArchetype(mask)
	mask.ForEachSet(func(id ComponentID) {
		info := r.componentsRegistry.idToInfo[id]
		slice := reflect.MakeSlice(reflect.SliceOf(info.Type), initCapacity, initCapacity)
		arch.columns[id] = &column{
			data:     slice.UnsafePointer(),
			dataType: info.Type,
			itemSize: info.Size,
			len:      0,
			cap:      initCapacity,
		}
	})

	r.archetypeMap[mask] = arch
	r.archetypes = append(r.archetypes, arch)
	return arch
}

func (r *archetypeRegistry) MoveEntity(entity Entity, oldArch *archetype, newArch *archetype, newCompID ComponentID, newData unsafe.Pointer) {
	if oldArch == nil {
		newArch.addEntity(entity, newCompID, newData)
		return
	}

	oldIdx, ok := oldArch.entityToIndex[entity]
	if !ok {
		newArch.addEntity(entity, newCompID, newData)
		return
	}

	newArch.ensureCapacity()
	newIdx := newArch.registerEntity(entity)

	for id, oldCol := range oldArch.columns {
		if newCol, exists := newArch.columns[id]; exists {
			src := oldCol.GetElement(oldIdx)
			newCol.setData(newIdx, src)
		}
	}

	if newCol, ok := newArch.columns[newCompID]; ok {
		newCol.setData(newIdx, newData)
	}

	oldArch.removeEntity(oldIdx)
}

func (r *archetypeRegistry) MoveEntityOnly(entity Entity, oldArch *archetype, newArch *archetype) {
	oldIdx, ok := oldArch.entityToIndex[entity]
	if !ok {
		return
	}

	newArch.ensureCapacity()
	newIdx := newArch.registerEntity(entity)

	for id, newCol := range newArch.columns {
		if oldCol, exists := oldArch.columns[id]; exists {
			src := oldCol.GetElement(oldIdx)
			newCol.setData(newIdx, src)
		}
	}

	oldArch.removeEntity(oldIdx)
}
