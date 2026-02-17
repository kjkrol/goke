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

func (r *ArchetypeRegistry) InitArchetype(
	mask ArchetypeMask,
	initCapacity int,
) ArchetypeId {
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
	arch.InitArchetype(archId, mask, r.sharedColsInfos)

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

func (r *ArchetypeRegistry) AddEntity(
	entity Entity,
	archId ArchetypeId,
) (ChunkIdx, ChunkRow) {
	arch := &r.Archetypes[archId]
	chunkIdx, chunkRow := arch.AddEntity(entity)
	r.EntityLinkStore.Update(entity, archId, chunkIdx, chunkRow)
	return chunkIdx, chunkRow
}

func (r *ArchetypeRegistry) UnlinkEntity(entity Entity) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}

	linkArch := &r.Archetypes[link.ArchId]
	swappedEntity, swapped := linkArch.SwapRemoveEntity(link.ChunkIdx, link.ChunkRow)

	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.ArchId, link.ChunkIdx, link.ChunkRow)
	}

	r.EntityLinkStore.Clear(entity)
}

var (
	ErrEntityNotFound = errors.New("entity not found in registry")
)

// AllocateComponentMemory ensures the entity has the component and
// returns a pointer to its memory.
func (r *ArchetypeRegistry) AllocateComponentMemory(
	entity Entity,
	compInfo ComponentInfo,
) (unsafe.Pointer, error) {
	compID := compInfo.ID

	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return nil, ErrEntityNotFound
	}

	currentArch := &r.Archetypes[link.ArchId]

	var targetArch *Archetype
	var targetChunkIdx ChunkIdx
	var targetRow ChunkRow

	// -------------------------------------------------------------------------
	// "FAST PATH" (Component already exists/allocated)
	// -------------------------------------------------------------------------
	if currentArch.Mask.IsSet(compID) {
		targetArch = currentArch
		targetChunkIdx = link.ChunkIdx
		targetRow = link.ChunkRow
	} else {

		// ---------------------------------------------------------------------
		// "SLOW PATH" (Transition to antother archetype)
		// ---------------------------------------------------------------------
		nextArchID := r.ensureNextEdgeId(compID, currentArch)
		targetChunkIdx, targetRow = r.moveEntity(entity, link, nextArchID)
		targetArch = &r.Archetypes[nextArchID]
	}

	// ignore "tags" - zero size components
	if compInfo.Size == 0 {
		return nil, nil
	}

	column := targetArch.GetColumn(compID)
	chunk := targetArch.Memory.GetChunk(targetChunkIdx)
	return column.GetPointer(chunk, targetRow), nil
}

func (r *ArchetypeRegistry) ensureNextEdgeId(
	compID ComponentID,
	oldArch *Archetype,
) ArchetypeId {
	// -------------------------------------------------------------------------
	// "FAST PATH" (Graph Edge Exists)
	// -------------------------------------------------------------------------
	if nextEdgeId := oldArch.graph.edgesNext[compID]; nextEdgeId != NullArchetypeId {
		return nextEdgeId
	}

	// -------------------------------------------------------------------------
	// "SLOW PATH" (Create new graph edge; create new arch if needed)
	// -------------------------------------------------------------------------
	newMask := oldArch.Mask.Set(compID)
	nextArchId := r.getOrRegister(newMask)

	// Create link in the graph
	nextArch := &r.Archetypes[nextArchId]
	oldArch.linkNextArch(nextArch, compID)

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

	// -------------------------------------------------------------------------
	// "FAST PATH" (Graph Edge Exists)
	// -------------------------------------------------------------------------
	if prevArchId := oldArch.graph.edgesPrev[compID]; prevArchId != NullArchetypeId {
		prevArch := &r.Archetypes[prevArchId]

		if prevArch.Mask.IsEmpty() {
			r.UnlinkEntity(entity)
			return
		}

		r.moveEntity(entity, link, prevArchId)
		return
	}

	// -------------------------------------------------------------------------
	// "SLOW PATH" (Calculate Mask & Create Edge)
	// -------------------------------------------------------------------------
	newMask := oldArch.Mask.Clear(compID)

	// Same mask? It means entity doesnt have component to remove.
	if oldArch.Mask == newMask {
		return
	}

	// Target mask is empty == Root Archetype; Lets remove entity without components.
	if newMask.IsEmpty() {
		r.UnlinkEntity(entity)
		return
	}

	// Find or create new target archetype
	newPrevArchId := r.getOrRegister(newMask)
	newPrevArch := &r.Archetypes[newPrevArchId]

	// Update Archetype Graf
	oldArch.linkPrevArch(newPrevArch, compID)

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

func (r *ArchetypeRegistry) moveEntity(
	entity Entity,
	link EntityArchLink,
	archId ArchetypeId,
) (ChunkIdx, ChunkRow) {
	oldArch := &r.Archetypes[link.ArchId]
	newArch := &r.Archetypes[archId]

	newChunkIdx, newRow := newArch.AddEntity(entity)

	srcChunk := oldArch.Memory.GetChunk(link.ChunkIdx)
	dstChunk := newArch.Memory.GetChunk(newChunkIdx)

	// itarate through new archetype columns
	for i := range newArch.Columns {
		dstCol := &newArch.Columns[i]

		if dstCol.CompID == EntityID {
			continue
		}

		// copying shared components
		if srcCol := oldArch.GetColumn(dstCol.CompID); srcCol != nil {
			srcPtr := srcCol.GetPointer(srcChunk, link.ChunkRow)
			dstPtr := dstCol.GetPointer(dstChunk, newRow)

			copyMemory(dstPtr, srcPtr, dstCol.ItemSize)
		}
	}

	// Remove from old location (Swap Remove)
	swappedEntity, swapped := oldArch.SwapRemoveEntity(link.ChunkIdx, link.ChunkRow)

	// Update LinkStore
	// If an entity was swapped to fill the gap -> update it
	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.ArchId, link.ChunkIdx, link.ChunkRow)
	}

	// Update the moved entity's location
	r.EntityLinkStore.Update(entity, newArch.Id, newChunkIdx, newRow)

	return newChunkIdx, newRow
}
