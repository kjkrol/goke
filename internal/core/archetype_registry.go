package core

import (
	"errors"
	"reflect"
	"unsafe"
)

type ArchetypeRegistry struct {
	archetypeMap              map[ArchetypeMask]*Archetype
	archetypes                []*Archetype
	EntityLinkStore           *EntityLinkStore
	componentsRegistry        *ComponentsRegistry
	viewRegistry              *ViewRegistry
	rootArch                  *Archetype
	defaultArchetypeChunkSize int
}

func NewArchetypeRegistry(
	componentsRegistry *ComponentsRegistry,
	viewRegistry *ViewRegistry,
	cfg RegistryConfig,
) *ArchetypeRegistry {
	reg := &ArchetypeRegistry{
		archetypeMap:              make(map[ArchetypeMask]*Archetype),
		archetypes:                make([]*Archetype, 0, cfg.InitialArchetypeRegistryCap),
		EntityLinkStore:           NewEntityLinkStore(cfg.InitialEntityCap),
		componentsRegistry:        componentsRegistry,
		viewRegistry:              viewRegistry,
		defaultArchetypeChunkSize: cfg.DefaultArchetypeChunkSize,
	}

	rootMask := ArchetypeMask{}
	reg.rootArch = NewArchetype(rootMask, reg.defaultArchetypeChunkSize)

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
	row := r.rootArch.registerEntity(entity)
	r.EntityLinkStore.Update(entity, r.rootArch, row)
}

func (r *ArchetypeRegistry) UnlinkEntity(entity Entity) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}

	swappedEntity, swapped := link.Arch.SwapRemoveEntity(link.Row)

	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.Arch, link.Row)
	}

	r.EntityLinkStore.Clear(entity)
}

var (
	ErrEntityNotFound = errors.New("entity not found in registry")
)

// AllocateComponentMemory ensures the entity has the component and returns a pointer to its memory.
// This avoids the 'escape to heap' allocation of the component data struct.
func (r *ArchetypeRegistry) AllocateComponentMemory(entity Entity, compInfo ComponentInfo) (unsafe.Pointer, error) {
	compID := compInfo.ID

	backLink, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return nil, ErrEntityNotFound
	}

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
		nextArch := oldArch.edgesNext[compID]
		if nextArch == nil {
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
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}
	oldArch := link.Arch
	compID := compInfo.ID

	// FAST PATH (use Archetype-Graph)
	if prevArch := oldArch.edgesPrev[compID]; prevArch != nil {
		if prevArch.Mask.IsEmpty() {
			r.UnlinkEntity(entity)
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
		r.UnlinkEntity(entity)
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

	arch := NewArchetype(mask, r.defaultArchetypeChunkSize)
	initCapacity := r.defaultArchetypeChunkSize

	for id := range mask.AllSet() {
		info := r.componentsRegistry.idToInfo[id]
		slice := reflect.MakeSlice(reflect.SliceOf(info.Type), initCapacity, initCapacity)
		arch.Columns[id] = &Column{
			Data:     slice.UnsafePointer(),
			dataType: info.Type,
			ItemSize: info.Size,
			len:      0,
			cap:      initCapacity,
		}
	}

	r.archetypeMap[mask] = arch
	r.archetypes = append(r.archetypes, arch)
	r.viewRegistry.OnArchetypeCreated(arch)
	return arch
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) moveEntity(entity Entity, link EntityArchLink, newArch *Archetype) ArchRow {
	oldArch := link.Arch
	oldArchRow := link.Row

	newArchRow := newArch.registerEntity(entity)

	for _, id := range newArch.activeIDs {
		if oldCol := oldArch.Columns[id]; oldCol != nil {
			newArch.Columns[id].setData(newArchRow, oldCol.GetElement(oldArchRow))
		}
	}

	r.UnlinkEntity(entity)

	r.EntityLinkStore.Update(entity, newArch, newArchRow)

	return newArchRow
}
