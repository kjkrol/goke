package ecs

import (
	"reflect"
	"unsafe"
)

const initArchetypesCapacity = 64

type archetypeRegistry struct {
	archetypeMap       map[ArchetypeMask]*Archetype
	archetypes         []*Archetype
	entityArchLinks    []EntityArchLink
	componentsRegistry *componentsRegistry
}

func newArchetypeRegistry(componentsRegistry *componentsRegistry) *archetypeRegistry {
	return &archetypeRegistry{
		archetypeMap:       make(map[ArchetypeMask]*Archetype),
		archetypes:         make([]*Archetype, 0, initArchetypesCapacity),
		entityArchLinks:    make([]EntityArchLink, 0, initialCapacity),
		componentsRegistry: componentsRegistry,
	}
}

func (r *archetypeRegistry) Get(mask ArchetypeMask) *Archetype {
	if arch, ok := r.archetypeMap[mask]; ok {
		return arch
	}
	return nil
}

func (r *archetypeRegistry) All() []*Archetype {
	return r.archetypes
}

func (r *archetypeRegistry) Assign(entity Entity, compID ComponentID, data unsafe.Pointer) {
	index := entity.Index()
	backLink := r.entityArchLinks[index]
	oldArch := backLink.arch

	if oldArch == nil {
		newMask := NewArchetypeMask(compID)
		newArch := r.getOrRegister(newMask)
		r.entityArchLinks[entity.Index()] = newArch.AddEntity(entity, compID, data)
		return
	}

	var oldMask = oldArch.mask
	newMask := oldMask.Set(compID)
	if oldMask == newMask {
		col := backLink.arch.columns[compID]
		col.setData(backLink.columnIndex, data)
		return
	}

	newArch := r.getOrRegister(newMask)
	r.moveEntity(entity, backLink, newArch, compID, data)
}

func (r *archetypeRegistry) UnAssign(entity Entity, compID ComponentID) {
	oldArch := r.entityArchLinks[entity.Index()].arch
	oldMask := oldArch.mask
	newMask := oldMask.Clear(compID)
	newArch := r.getOrRegister(newMask)
	r.moveEntityOnly(entity, oldArch, newArch)
}

func (r *archetypeRegistry) AddEntity(entity Entity) {
	index := entity.Index()
	for int(index) >= len(r.entityArchLinks) {
		r.entityArchLinks = append(r.entityArchLinks, EntityArchLink{})
	}
	r.entityArchLinks[index] = EntityArchLink{}
}

func (r *archetypeRegistry) RemoveEntity(entity Entity) {
	index := entity.Index()
	entityArchIndex := r.entityArchLinks[index].columnIndex
	r.entityArchLinks[entity].arch.RemoveEntity(entityArchIndex)
	r.entityArchLinks[entity] = EntityArchLink{}
}

// --------------------------------------------------------------

func (r *archetypeRegistry) getOrRegister(mask ArchetypeMask) *Archetype {
	if arch, ok := r.archetypeMap[mask]; ok {
		return arch
	}

	arch := NewArchetype(mask)
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

// --------------------------------------------------------------

func (r *archetypeRegistry) moveEntity(entity Entity, backLink EntityArchLink, newArch *Archetype, compID ComponentID, newData unsafe.Pointer) {
	oldArch := backLink.arch
	oldColumnIndex := backLink.columnIndex

	newColumnIndex := newArch.registerEntity(entity)

	for id, newCol := range newArch.columns {
		if oldCol, exists := oldArch.columns[id]; exists {
			newCol.setData(newColumnIndex, oldCol.GetElement(oldColumnIndex))
		} else if id == compID {
			newCol.setData(newColumnIndex, newData)
		} else {
			newCol.zeroData(newColumnIndex)
		}
	}

	// Swap-and-Pop
	swappedEntity, swaped := oldArch.RemoveEntity(oldColumnIndex)
	if swaped {
		r.entityArchLinks[swappedEntity.Index()].arch = oldArch
		r.entityArchLinks[swappedEntity.Index()].columnIndex = oldColumnIndex
	}

	r.entityArchLinks[entity.Index()].arch = newArch
	r.entityArchLinks[entity.Index()].columnIndex = newColumnIndex
}

func (r *archetypeRegistry) moveEntityOnly(entity Entity, oldArch *Archetype, newArch *Archetype) {
	index := entity.Index()
	oldColumnIndex := r.entityArchLinks[index].columnIndex

	newColumnIndex := newArch.registerEntity(entity)

	for id, newCol := range newArch.columns {
		if oldCol, exists := oldArch.columns[id]; exists {
			src := oldCol.GetElement(oldColumnIndex)
			newCol.setData(newColumnIndex, src)
		}
	}

	oldArch.RemoveEntity(oldColumnIndex)

	r.entityArchLinks[index].arch = newArch
	r.entityArchLinks[index].columnIndex = newColumnIndex
}
