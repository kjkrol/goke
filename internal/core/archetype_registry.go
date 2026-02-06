package core

import (
	"errors"
	"math/bits"
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
	reg.rootArch = &Archetype{}
	reg.InitArchetype(reg.rootArch, rootMask, reg.defaultArchetypeChunkSize)

	reg.archetypeMap[rootMask] = reg.rootArch
	reg.archetypes = append(reg.archetypes, reg.rootArch)

	return reg
}

func (r *ArchetypeRegistry) InitArchetype(arch *Archetype, mask ArchetypeMask, defaultArchetypeChunkSize int) {
	// Pre-calculate active IDs to avoid bitmask scanning in hot loops
	var activeIDs [MaxComponents]ComponentID
	counter := 0
	for i, word := range mask {
		for word != 0 {
			bitPos := bits.TrailingZeros64(word)
			id := ComponentID(i*64 + bitPos)
			if r.componentsRegistry.idToInfo[id].Size > 0 {
				activeIDs[counter] = id
				counter++
			}
			word &= word - 1
		}
	}

	*arch = Archetype{
		Mask:      mask,
		entities:  make([]Entity, defaultArchetypeChunkSize),
		activeIDs: activeIDs[:counter],
		len:       0,
		cap:       defaultArchetypeChunkSize,
		initCap:   defaultArchetypeChunkSize,
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
	r.addEntity(entity, r.rootArch)
}

func (r *ArchetypeRegistry) addEntity(entity Entity, arch *Archetype) ArchRow {
	row := arch.registerEntity(entity)
	r.EntityLinkStore.Update(entity, arch, row)
	return row
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

	currentArch := backLink.Arch
	var targetRow ArchRow
	var targetArch *Archetype

	// 1. If component already exists, just return the address
	if currentArch.Mask.IsSet(compID) {
		targetRow = backLink.Row
		targetArch = currentArch
	} else {
		// 2. Perform structural change (Archetype Transition)
		// Check if we have a fast path in the Archetype-Graph
		r.ensureNextEdge(compID, currentArch, &targetArch)
		// Move existing data to the new archetype
		targetRow = r.moveEntity(entity, backLink, targetArch)
	}

	if compInfo.Size == 0 {
		return nil, nil
	}

	// 3. Calculate and return the direct pointer
	column := targetArch.Columns[compID]
	return unsafe.Add(column.Data, uintptr(targetRow)*column.ItemSize), nil
}

func (r *ArchetypeRegistry) ensureNextEdge(compID ComponentID, oldArch *Archetype, nextArch **Archetype) {
	if nextEdge := oldArch.edgesNext[compID]; nextEdge != nil {
		*nextArch = nextEdge
		return
	}

	// Slow path: create or get new archetype
	newMask := oldArch.Mask.Set(compID)
	r.getOrRegister(newMask, nextArch)

	// Link in the graph
	actualNext := *nextArch
	oldArch.edgesNext[compID] = actualNext
	actualNext.edgesPrev[compID] = oldArch
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

	var newArch *Archetype = &Archetype{}
	r.getOrRegister(newMask, &newArch)

	// register edges on Archetype-Graph
	oldArch.edgesPrev[compID] = newArch
	newArch.edgesNext[compID] = oldArch

	oldArch.Columns[compID].zeroData(link.Row)
	r.moveEntity(entity, link, newArch)
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) getOrRegister(mask ArchetypeMask, arch **Archetype) {
	if found, ok := r.archetypeMap[mask]; ok {
		*arch = found
		return
	}

	*arch = &Archetype{}
	actualArch := *arch
	r.InitArchetype(actualArch, mask, r.defaultArchetypeChunkSize)
	initCapacity := r.defaultArchetypeChunkSize

	for id := range mask.AllSet() {
		info := r.componentsRegistry.idToInfo[id]
		// tags should not have columns
		if info.Size == 0 {
			continue
		}
		slice := reflect.MakeSlice(reflect.SliceOf(info.Type), initCapacity, initCapacity)
		actualArch.Columns[id] = &Column{
			Data:     slice.UnsafePointer(),
			dataType: info.Type,
			ItemSize: info.Size,
			len:      0,
			cap:      initCapacity,
		}
	}

	r.archetypeMap[mask] = actualArch
	r.archetypes = append(r.archetypes, actualArch)
	r.viewRegistry.OnArchetypeCreated(actualArch)
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
