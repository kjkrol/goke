package ecs

import (
	"errors"
	"reflect"
	"unsafe"
)

type ArchetypeRegistry struct {
	archetypeMap       map[ArchetypeMask]*Archetype
	archetypes         []*Archetype
	entityArchLinks    []EntityArchLink
	componentsRegistry *ComponentsRegistry
	viewRegistry       *ViewRegistry
	rootArch           *Archetype
}

func NewArchetypeRegistry(componentsRegistry *ComponentsRegistry, viewRegistry *ViewRegistry) *ArchetypeRegistry {
	reg := &ArchetypeRegistry{
		archetypeMap:       make(map[ArchetypeMask]*Archetype),
		archetypes:         make([]*Archetype, 0, archetypesInitCap),
		entityArchLinks:    make([]EntityArchLink, 0, entityPoolInitCap),
		componentsRegistry: componentsRegistry,
		viewRegistry:       viewRegistry,
	}

	rootMask := ArchetypeMask{}
	reg.rootArch = NewArchetype(rootMask)

	reg.archetypeMap[rootMask] = reg.rootArch
	reg.archetypes = append(reg.archetypes, reg.rootArch)

	return reg
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
	row := r.rootArch.registerEntity(entity)
	r.entityArchLinks[index] = EntityArchLink{
		arch: r.rootArch,
		row:  row,
	}
}

func (r *ArchetypeRegistry) RemoveEntity(entity Entity) {
	index := entity.Index()
	link := r.entityArchLinks[index]
	if link.arch == nil {
		return
	}

	swappedEntity, swapped := link.arch.SwapRemoveEntity(link.row)

	if swapped {
		r.entityArchLinks[swappedEntity.Index()].row = link.row
	}

	r.entityArchLinks[index] = EntityArchLink{arch: nil, row: 0}
}

var (
	ErrNilComponentData = errors.New("component data pointer cannot be nil")
	ErrEntityNotFound   = errors.New("entity not found in registry")
)

func (r *ArchetypeRegistry) Assign(entity Entity, compInfo ComponentInfo, data unsafe.Pointer) error {
	if data == nil && compInfo.Size > 0 {
		return ErrNilComponentData
	}

	compID := compInfo.ID

	index := entity.Index()
	if int(index) >= len(r.entityArchLinks) || r.entityArchLinks[index].arch == nil {
		return ErrEntityNotFound
	}
	backLink := r.entityArchLinks[index]
	oldArch := backLink.arch

	// override existing component
	if oldArch.mask.IsSet(compID) {
		oldArch.columns[compID].setData(backLink.row, data)
		return nil
	}

	// FAST PATH (use Archetype-Graph)
	if nextArch, ok := oldArch.edgesNext[compID]; ok {
		newArchRow := r.moveEntity(entity, backLink, nextArch)
		nextArch.columns[compID].setData(newArchRow, data)
		return nil
	}

	// SLOW PATH (nextArch does not exist yet)

	// move to another archetype
	newMask := oldArch.mask.Set(compID)
	newArch := r.getOrRegister(newMask)

	// register edges on Archetype-Graph
	oldArch.edgesNext[compID] = newArch
	newArch.edgesPrev[compID] = oldArch

	newArchRow := r.moveEntity(entity, backLink, newArch)
	newArch.columns[compID].setData(newArchRow, data)

	return nil
}

// AllocateComponentMemory ensures the entity has the component and returns a pointer to its memory.
// This avoids the 'escape to heap' allocation of the component data struct.
func (r *ArchetypeRegistry) AllocateComponentMemory(entity Entity, compInfo ComponentInfo) (unsafe.Pointer, error) {
	compID := compInfo.ID
	index := entity.Index()

	// Safety check (similar to your Assign)
	if int(index) >= len(r.entityArchLinks) || r.entityArchLinks[index].arch == nil {
		return nil, ErrEntityNotFound
	}

	backLink := r.entityArchLinks[index]
	oldArch := backLink.arch
	var targetRow ArchRow
	var targetArch *Archetype

	// 1. If component already exists, just return the address
	if oldArch.mask.IsSet(compID) {
		targetRow = backLink.row
		targetArch = oldArch
	} else {
		// 2. Perform structural change (Archetype Transition)
		// Check if we have a fast path in the Archetype-Graph
		nextArch, ok := oldArch.edgesNext[compID]
		if !ok {
			// Slow path: create or get new archetype
			newMask := oldArch.mask.Set(compID)
			nextArch = r.getOrRegister(newMask)

			// Link in the graph
			oldArch.edgesNext[compID] = nextArch
			nextArch.edgesPrev[compID] = oldArch
		}

		// Move existing data to the new archetype
		targetRow = r.moveEntity(entity, backLink, nextArch)
		targetArch = nextArch
	}

	// 3. Calculate and return the direct pointer
	column := targetArch.columns[compID]
	return unsafe.Add(column.data, uintptr(targetRow)*column.itemSize), nil
}

func (r *ArchetypeRegistry) UnAssign(entity Entity, compInfo ComponentInfo) {
	index := entity.Index()
	link := r.entityArchLinks[index]
	oldArch := link.arch
	compID := compInfo.ID

	// FAST PATH (use Archetype-Graph)
	if prevArch, ok := oldArch.edgesPrev[compID]; ok {
		if prevArch.mask.IsEmpty() {
			r.RemoveEntity(entity)
			return
		}
		oldArch.columns[compID].zeroData(link.row)
		r.moveEntity(entity, link, prevArch)
		return
	}
	// SLOW PATH (prevArch does not exist yet)
	newMask := oldArch.mask.Clear(compID)

	// nothing to unassign
	if oldArch.mask == newMask {
		return
	}

	if newMask.IsEmpty() {
		r.RemoveEntity(entity)
		return
	}

	newArch := r.getOrRegister(newMask)

	// register edges on Archetype-Graph
	oldArch.edgesPrev[compID] = newArch
	newArch.edgesNext[compID] = oldArch

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
