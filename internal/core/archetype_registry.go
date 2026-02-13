package core

import (
	"errors"
	"fmt"
	"unsafe"
)

type ArchetypeRegistry struct {
	archetypeMaskMap          ArchetypeMaskMap
	Archetypes                [MaxArchetypeId]Archetype
	lastArchetypeId           ArchetypeId
	EntityLinkStore           EntityLinkStore
	componentsRegistry        *ComponentsRegistry
	viewRegistry              *ViewRegistry
	defaultArchetypeChunkSize int
	sharedColsInfos           []ComponentInfo
}

func NewArchetypeRegistry(
	componentsRegistry *ComponentsRegistry,
	viewRegistry *ViewRegistry,
	cfg RegistryConfig,
) *ArchetypeRegistry {
	reg := ArchetypeRegistry{
		EntityLinkStore:           NewEntityLinkStore(cfg.InitialEntityCap),
		componentsRegistry:        componentsRegistry,
		viewRegistry:              viewRegistry,
		defaultArchetypeChunkSize: cfg.DefaultArchetypeChunkSize,
		lastArchetypeId:           RootArchetypeId,
	}

	rootMask := ArchetypeMask{}
	reg.InitArchetype(rootMask, reg.defaultArchetypeChunkSize)

	return &reg
}

func (r *ArchetypeRegistry) InitArchetype(mask ArchetypeMask, initCapacity int) ArchetypeId {
	if r.lastArchetypeId >= MaxArchetypeId {
		panic(fmt.Sprintf("Max archetype number exceeded: %d", MaxArchetypeId))
	}

	r.sharedColsInfos = r.sharedColsInfos[:0]

	defer func() {
		clear(r.sharedColsInfos)
		r.sharedColsInfos = r.sharedColsInfos[:0]
	}()

	for id := range mask.AllSet() {
		info := r.componentsRegistry.idToInfo[id]
		// tags should not have columns
		if info.Size == 0 {
			continue
		}
		r.sharedColsInfos = append(r.sharedColsInfos, info)
	}

	archId := r.lastArchetypeId
	arch := &r.Archetypes[archId]
	arch.InitArchetype(archId, mask, r.sharedColsInfos, initCapacity)

	r.archetypeMaskMap.Put(mask, archId)

	r.lastArchetypeId++

	return archId
}

func (r *ArchetypeRegistry) Get(mask ArchetypeMask) ArchetypeId {
	if archId, ok := r.archetypeMaskMap.Get(mask); ok {
		return archId
	}
	return NullArchetypeId
}

func (r *ArchetypeRegistry) AddEntity(entity Entity, archId ArchetypeId) ArchRow {
	arch := &r.Archetypes[archId]
	row := arch.registerEntity(entity)
	r.EntityLinkStore.Update(entity, archId, row)
	return row
}

func (r *ArchetypeRegistry) UnlinkEntity(entity Entity) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}

	linkArch := &r.Archetypes[link.ArchId]
	swappedEntity, swapped := linkArch.SwapRemoveEntity(link.Row)

	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.ArchId, link.Row)
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

	currentArch := &r.Archetypes[backLink.ArchId]
	var targetRow ArchRow
	targetArchId := NullArchetypeId

	// 1. If component already exists, just return the address
	if currentArch.Mask.IsSet(compID) {
		targetRow = backLink.Row
		targetArchId = currentArch.Id
	} else {
		// 2. Perform structural change (Archetype Transition)
		// Check if we have a fast path in the Archetype-Graph
		targetArchId = r.ensureNextEdgeId(compID, currentArch)
		// Move existing data to the new archetype
		targetRow = r.moveEntity(entity, backLink, targetArchId)
	}

	if compInfo.Size == 0 {
		return nil, nil
	}

	// 3. Calculate and return the direct pointer
	column := r.Archetypes[targetArchId].GetColumn(compID)
	return unsafe.Add(column.Data, uintptr(targetRow)*column.ItemSize), nil
}

func (r *ArchetypeRegistry) ensureNextEdgeId(compID ComponentID, oldArch *Archetype) ArchetypeId {
	if nextEdgeId := oldArch.graph.edgesNext[compID]; nextEdgeId != NullArchetypeId {
		return nextEdgeId
	}

	// Slow path: create or get new archetype
	newMask := oldArch.Mask.Set(compID)
	nextArchId := r.getOrRegister(newMask)

	// Link in the graph
	actualNext := &r.Archetypes[nextArchId]
	oldArch.graph.edgesNext[compID] = actualNext.Id
	actualNext.graph.edgesPrev[compID] = oldArch.Id
	return nextArchId
}

func (r *ArchetypeRegistry) UnAssign(entity Entity, compInfo ComponentInfo) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}
	oldArchId := link.ArchId
	compID := compInfo.ID

	oldArch := &r.Archetypes[oldArchId]

	// FAST PATH (use Archetype-Graph)
	if prevArchId := oldArch.graph.edgesPrev[compID]; prevArchId != NullArchetypeId {
		prevArch := &r.Archetypes[prevArchId]
		if prevArch.Mask.IsEmpty() {
			r.UnlinkEntity(entity)
			return
		}
		oldArch.GetColumn(compID).ZeroData(link.Row)
		r.moveEntity(entity, link, prevArchId)
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

	newPrevArchId := r.getOrRegister(newMask)

	newPrevArch := &r.Archetypes[newPrevArchId]
	// register edges on Archetype-Graph
	oldArch.graph.edgesPrev[compID] = newPrevArchId
	newPrevArch.graph.edgesNext[compID] = oldArch.Id

	oldArch.GetColumn(compID).ZeroData(link.Row)
	r.moveEntity(entity, link, newPrevArchId)
}

func (r *ArchetypeRegistry) Reset() {
	for i := range int(r.lastArchetypeId) {
		r.Archetypes[i].Reset()
	}
	clear(r.Archetypes[:])

	r.archetypeMaskMap.Reset()
	r.EntityLinkStore.Reset()

	clear(r.sharedColsInfos)
	r.sharedColsInfos = r.sharedColsInfos[:0]

	r.lastArchetypeId = RootArchetypeId
	rootMask := ArchetypeMask{}
	r.InitArchetype(rootMask, r.defaultArchetypeChunkSize)
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) getOrRegister(mask ArchetypeMask) ArchetypeId {
	if found, ok := r.archetypeMaskMap.Get(mask); ok {
		return found
	}
	archId := r.InitArchetype(mask, r.defaultArchetypeChunkSize)
	r.viewRegistry.OnArchetypeCreated(&r.Archetypes[archId])
	return archId
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) moveEntity(entity Entity, link EntityArchLink, archId ArchetypeId) ArchRow {
	oldArch := &r.Archetypes[link.ArchId]
	oldArchRow := link.Row

	newArch := &r.Archetypes[archId]
	newArchRow := newArch.registerEntity(entity)

	for i := int(FirstDataColumnIndex); i < len(newArch.block.Columns); i++ {
		col := &newArch.block.Columns[i]
		if oldCol := oldArch.GetColumn(col.CompID); oldCol != nil {
			col.SetData(newArchRow, oldCol.GetElement(oldArchRow))
		}
	}

	swappedEntity, swapped := oldArch.SwapRemoveEntity(link.Row)

	if swapped {
		r.EntityLinkStore.Update(swappedEntity, oldArch.Id, link.Row)
	}

	r.EntityLinkStore.Update(entity, newArch.Id, newArchRow)

	return newArchRow
}
