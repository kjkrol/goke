package core

import (
	"errors"
	"reflect"
	"unsafe"
)

type ArchetypeRegistry struct {
	archetypeMap       map[ArchetypeMask]*Archetype
	archetypes         []*Archetype
	EntityArchLinks    []EntityArchLink
	componentsRegistry *ComponentsRegistry
	viewRegistry       *ViewRegistry
	rootArch           *Archetype
}

func NewArchetypeRegistry(componentsRegistry *ComponentsRegistry, viewRegistry *ViewRegistry) *ArchetypeRegistry {
	reg := &ArchetypeRegistry{
		archetypeMap:       make(map[ArchetypeMask]*Archetype),
		archetypes:         make([]*Archetype, 0, archetypesInitCap),
		EntityArchLinks:    make([]EntityArchLink, 0, entityPoolInitCap),
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
	for int(index) >= len(r.EntityArchLinks) {
		r.EntityArchLinks = append(r.EntityArchLinks, EntityArchLink{})
	}
	row := r.rootArch.registerEntity(entity)
	r.EntityArchLinks[index] = EntityArchLink{
		Arch: r.rootArch,
		Row:  row,
	}
}

func (r *ArchetypeRegistry) RemoveEntity(entity Entity) {
	index := entity.Index()
	link := r.EntityArchLinks[index]
	if link.Arch == nil {
		return
	}

	swappedEntity, swapped := link.Arch.SwapRemoveEntity(link.Row)

	if swapped {
		r.EntityArchLinks[swappedEntity.Index()].Row = link.Row
	}

	r.EntityArchLinks[index] = EntityArchLink{Arch: nil, Row: 0}
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
	if int(index) >= len(r.EntityArchLinks) || r.EntityArchLinks[index].Arch == nil {
		return ErrEntityNotFound
	}
	backLink := r.EntityArchLinks[index]
	oldArch := backLink.Arch

	// override existing component
	if oldArch.Mask.IsSet(compID) {
		oldArch.Columns[compID].setData(backLink.Row, data)
		return nil
	}

	// FAST PATH (use Archetype-Graph)
	if nextArch, ok := oldArch.edgesNext[compID]; ok {
		newArchRow := r.moveEntity(entity, backLink, nextArch)
		nextArch.Columns[compID].setData(newArchRow, data)
		return nil
	}

	// SLOW PATH (nextArch does not exist yet)

	// move to another archetype
	newMask := oldArch.Mask.Set(compID)
	newArch := r.getOrRegister(newMask)

	// register edges on Archetype-Graph
	oldArch.edgesNext[compID] = newArch
	newArch.edgesPrev[compID] = oldArch

	newArchRow := r.moveEntity(entity, backLink, newArch)
	newArch.Columns[compID].setData(newArchRow, data)

	return nil
}

// AllocateComponentMemory ensures the entity has the component and returns a pointer to its memory.
// This avoids the 'escape to heap' allocation of the component data struct.
func (r *ArchetypeRegistry) AllocateComponentMemory(entity Entity, compInfo ComponentInfo) (unsafe.Pointer, error) {
	compID := compInfo.ID
	index := entity.Index()

	// Safety check (similar to your Assign)
	if int(index) >= len(r.EntityArchLinks) || r.EntityArchLinks[index].Arch == nil {
		return nil, ErrEntityNotFound
	}

	backLink := r.EntityArchLinks[index]
	oldArch := backLink.Arch
	var targetRow ArchRow
	var targetArch *Archetype

	// 1. If component already exists, just return the address
	if oldArch.Mask.IsSet(compID) {
		targetRow = backLink.Row
		targetArch = oldArch
	} else {
		// 2. Perform structural change (Archetype Transition)
		// Check if we have a fast path in the Archetype-Graph
		nextArch, ok := oldArch.edgesNext[compID]
		if !ok {
			// Slow path: create or get new archetype
			newMask := oldArch.Mask.Set(compID)
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
	column := targetArch.Columns[compID]
	return unsafe.Add(column.Data, uintptr(targetRow)*column.ItemSize), nil
}

func (r *ArchetypeRegistry) UnAssign(entity Entity, compInfo ComponentInfo) {
	index := entity.Index()
	link := r.EntityArchLinks[index]
	oldArch := link.Arch
	compID := compInfo.ID

	// FAST PATH (use Archetype-Graph)
	if prevArch, ok := oldArch.edgesPrev[compID]; ok {
		if prevArch.Mask.IsEmpty() {
			r.RemoveEntity(entity)
			return
		}
		oldArch.Columns[compID].zeroData(link.Row)
		r.moveEntity(entity, link, prevArch)
		return
	}
	// SLOW PATH (prevArch does not exist yet)
	newMask := oldArch.Mask.Clear(compID)

	// nothing to unassign
	if oldArch.Mask == newMask {
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

	oldArch.Columns[compID].zeroData(link.Row)
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
		arch.Columns[id] = &Column{
			Data:     slice.UnsafePointer(),
			dataType: info.Type,
			ItemSize: info.Size,
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
	oldArch := link.Arch
	oldArchRow := link.Row

	newArchRow := newArch.registerEntity(entity)

	for id, newCol := range newArch.Columns {
		if oldCol, exists := oldArch.Columns[id]; exists {
			newCol.setData(newArchRow, oldCol.GetElement(oldArchRow))
		}
	}

	r.RemoveEntity(entity)

	r.EntityArchLinks[index].Arch = newArch
	r.EntityArchLinks[index].Row = newArchRow

	return newArchRow
}
