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

func (r *archetypeRegistry) MoveEntity(reg *Registry, entity Entity, oldArch *archetype, newArch *archetype, newCompID ComponentID, newData unsafe.Pointer) {
	entID := uint32(entity & IndexMask)

	// Scenario A: initial component
	if oldArch == nil {
		newIdx := newArch.addEntity(entity, newCompID, newData)
		reg.entitiesRegistry.SetBacklink(entity, newArch, newIdx)
		return
	}

	// Scenario B: migration between archetypes
	oldIdx := reg.entitiesRegistry.records[entID].index
	newArch.ensureCapacity()
	newIdx := newArch.registerEntity(entity)

	for id, newCol := range newArch.columns {
		if oldCol, exists := oldArch.columns[id]; exists {
			newCol.setData(newIdx, oldCol.GetElement(oldIdx))
		} else if id == newCompID {
			newCol.setData(newIdx, newData)
		} else {
			newCol.zeroData(newIdx)
		}
	}

	// Swap-and-Pop
	swappedEntity := oldArch.removeEntity(oldIdx)
	if swappedEntity != 0 {
		reg.entitiesRegistry.SetBacklink(swappedEntity, oldArch, oldIdx)
	}

	reg.entitiesRegistry.SetBacklink(entity, newArch, newIdx)
}

func (r *archetypeRegistry) MoveEntityOnly(reg *Registry, entity Entity, oldArch *archetype, newArch *archetype) {
	rec, ok := reg.entitiesRegistry.GetRecord(entity)
	if !ok {
		return
	}
	oldIdx := rec.index

	newArch.ensureCapacity()
	newIdx := newArch.registerEntity(entity)

	for id, newCol := range newArch.columns {
		if oldCol, exists := oldArch.columns[id]; exists {
			src := oldCol.GetElement(oldIdx)
			newCol.setData(newIdx, src)
		}
	}

	oldArch.removeEntity(oldIdx)

	rec.arch = newArch
	rec.index = newIdx
}
