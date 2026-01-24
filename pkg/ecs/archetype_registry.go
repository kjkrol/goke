package ecs

import (
	"reflect"
	"unsafe"
)

const initArchetypesCapacity = 64

type ArchetypeRegistry struct {
	archetypeMap       map[ArchetypeMask]*Archetype
	archetypes         []*Archetype
	entityArchLinks    []EntityArchLink
	componentsRegistry *ComponentsRegistry
	viewRegistry       *ViewRegistry
}

func newArchetypeRegistry(
	componentsRegistry *ComponentsRegistry,
	viewRegistry *ViewRegistry,
) *ArchetypeRegistry {
	return &ArchetypeRegistry{
		archetypeMap:       make(map[ArchetypeMask]*Archetype),
		archetypes:         make([]*Archetype, 0, initArchetypesCapacity),
		entityArchLinks:    make([]EntityArchLink, 0, initialCapacity),
		componentsRegistry: componentsRegistry,
		viewRegistry:       viewRegistry,
	}
}

func (r *ArchetypeRegistry) Get(mask ArchetypeMask) *Archetype {
	if arch, ok := r.archetypeMap[mask]; ok {
		return arch
	}
	return nil
}

func (r *ArchetypeRegistry) All() []*Archetype {
	return r.archetypes
}

func (r *ArchetypeRegistry) AddEntity(entity Entity) {
	index := entity.Index()
	for int(index) >= len(r.entityArchLinks) {
		r.entityArchLinks = append(r.entityArchLinks, EntityArchLink{})
	}
	r.entityArchLinks[index] = EntityArchLink{}
}

func (r *ArchetypeRegistry) RemoveEntity(entity Entity) {
	index := entity.Index()
	link := r.entityArchLinks[index]
	swappedEntity, swaped := link.arch.SwapRemoveEntity(link.row)

	// Swap-and-Pop
	if swaped {
		// RemoveEntity moved last item of Archetype.entites to oldLink.columnIndex, so we have to update entityArchLinks
		r.entityArchLinks[swappedEntity.Index()].arch = link.arch
		r.entityArchLinks[swappedEntity.Index()].row = link.row
	}
}

func (r *ArchetypeRegistry) Assign(entity Entity, compID ComponentID, data unsafe.Pointer) {
	index := entity.Index()
	backLink := r.entityArchLinks[index]
	oldArch := backLink.arch

	// first time register
	if oldArch == nil {
		newMask := NewArchetypeMask(compID)
		newArch := r.getOrRegister(newMask)
		r.entityArchLinks[index] = newArch.AddEntity(entity, compID, data)
		return
	}

	var oldMask = oldArch.mask
	newMask := oldMask.Set(compID)

	// override existing component
	if oldMask == newMask {
		col := backLink.arch.columns[compID]
		col.setData(backLink.row, data)
		return
	}

	// move to another archetype
	newArch := r.getOrRegister(newMask)
	newArchRow := r.moveEntity(entity, backLink, newArch)
	newArch.columns[compID].setData(newArchRow, data)
}

func (r *ArchetypeRegistry) UnAssign(entity Entity, compID ComponentID) {
	link := r.entityArchLinks[entity.Index()]
	oldArch := link.arch
	oldMask := oldArch.mask
	newMask := oldMask.Clear(compID)

	// nothing to unassign
	if oldMask == newMask {
		return
	}

	if newMask.IsEmpty() {
		r.RemoveEntity(entity)
		return
	}

	newArch := r.getOrRegister(newMask)
	oldArch.columns[compID].zeroData(link.row)
	r.moveEntity(entity, link, newArch)
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) getOrRegister(mask ArchetypeMask) *Archetype {
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
	r.viewRegistry.OnArchetypeCreated(arch)
	return arch
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) moveEntity(entity Entity, link EntityArchLink, newArch *Archetype) ArchRow {
	index := entity.Index()
	oldArch := link.arch
	oldArchRow := link.row

	newArchRow := newArch.registerEntity(entity)

	for id, newCol := range newArch.columns {
		if oldCol, exists := oldArch.columns[id]; exists {
			newCol.setData(newArchRow, oldCol.GetElement(oldArchRow))
		}
	}

	r.RemoveEntity(entity)

	r.entityArchLinks[index].arch = newArch
	r.entityArchLinks[index].row = newArchRow

	return newArchRow
}
