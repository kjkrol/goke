package arch

import (
	"fmt"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/soa"
)

type Catalog struct {
	maskIndex       MaskIndex
	Archetypes      [MaxID]Archetype
	lastArchetypeId ID
	EntityIndex     EntityIndex
	observer        Observer
}

func (r *Catalog) Init(observer Observer, initialEntityCap int) {
	r.EntityIndex = NewEntityIndex(initialEntityCap)
	r.observer = observer
	r.lastArchetypeId = RootID
	r.addArchetype(comp.Composition{})
}

func (r *Catalog) addArchetype(set comp.Composition) ID {
	if r.lastArchetypeId >= MaxID {
		panic(fmt.Sprintf("Max archetype number exceeded: %d", MaxID))
	}
	archId := r.lastArchetypeId
	r.Archetypes[archId].Init(archId, set)
	r.maskIndex.Upsert(set.Mask, archId)
	r.lastArchetypeId++
	return archId
}

func (r *Catalog) Len() ID {
	return r.lastArchetypeId
}

func (r *Catalog) Upsert(set comp.Composition) ID {
	if found, ok := r.maskIndex.Get(set.Mask); ok {
		return found
	}
	archId := r.addArchetype(set)
	r.observer.OnArchetypeCreated(&r.Archetypes[archId])
	return archId
}

// TODO: metody ponizej dotycza zarzadzania encjami, nie archetypami — do przemyslenia czy naleza do Catalog

func (r *Catalog) AddEntity(entityID uid.UID64, archId ID) soa.BlockPos {
	arch := &r.Archetypes[archId]
	pos := arch.Table.AddEntity(entityID)
	r.EntityIndex.Upsert(entityID, archId, pos)
	return pos
}

func (r *Catalog) UnlinkEntity(entityID uid.UID64) {
	link, ok := r.EntityIndex.Get(entityID)
	if !ok {
		return
	}
	linkArch := &r.Archetypes[link.ArchId]
	swappedEntity, swapped := linkArch.Table.SwapRemoveEntity(link.Pos)
	if swapped {
		r.EntityIndex.Upsert(swappedEntity, link.ArchId, link.Pos)
	}
	r.EntityIndex.Clear(entityID)
}

func (r *Catalog) UpsertComp(entityID uid.UID64, compMeta comp.Meta) (unsafe.Pointer, error) {
	compID := compMeta.ID

	link, ok := r.EntityIndex.Get(entityID)
	if !ok {
		panic("arch: UpsertComp called with entityID not present in EntityIndex — broken reg invariant")
	}

	currentArch := &r.Archetypes[link.ArchId]

	var targetArch *Archetype
	var targetPos soa.BlockPos

	if currentArch.set.Mask.IsSet(compID) {
		targetArch = currentArch
		targetPos = link.Pos
	} else {
		nextArchID := r.ensureNextEdgeId(compMeta, currentArch)
		targetPos = r.moveEntity(entityID, link, nextArchID)
		targetArch = &r.Archetypes[nextArchID]
	}

	if compMeta.Size == 0 {
		return nil, nil
	}

	column := targetArch.Table.GetColumn(compID)
	chunk := targetArch.Table.GetChunk(targetPos.ChunkIdx)
	return column.At(chunk, targetPos.ChunkSlot), nil
}

func (r *Catalog) ensureNextEdgeId(compMeta comp.Meta, oldArch *Archetype) ID {
	compID := compMeta.ID
	if nextEdgeId := oldArch.graph.edgesNext[compID]; nextEdgeId != NullID {
		return nextEdgeId
	}

	nextArchId := r.Upsert(oldArch.set.With(compMeta))

	nextArch := &r.Archetypes[nextArchId]
	oldArch.graph.linkNext(nextArch.graph, oldArch.Id, nextArch.Id, compID)

	return nextArchId
}

func (r *Catalog) RemoveComp(entityID uid.UID64, compMeta comp.Meta) {
	link, ok := r.EntityIndex.Get(entityID)
	if !ok {
		return
	}

	compID := compMeta.ID
	oldArch := &r.Archetypes[link.ArchId]

	if !oldArch.set.Mask.IsSet(compID) {
		return
	}

	if prevArchId := oldArch.graph.edgesPrev[compID]; prevArchId != NullID {
		prevArch := &r.Archetypes[prevArchId]
		if prevArch.set.Mask.IsEmpty() {
			r.UnlinkEntity(entityID)
			return
		}
		r.moveEntity(entityID, link, prevArchId)
		return
	}

	newSet := oldArch.set.Without(compID)

	if newSet.Mask.IsEmpty() {
		r.UnlinkEntity(entityID)
		return
	}

	newPrevArchId := r.Upsert(newSet)
	newPrevArch := &r.Archetypes[newPrevArchId]
	oldArch.graph.linkPrev(newPrevArch.graph, oldArch.Id, newPrevArch.Id, compID)

	r.moveEntity(entityID, link, newPrevArchId)
}

func (r *Catalog) moveEntity(entityID uid.UID64, link EntityLocation, archId ID) soa.BlockPos {
	oldArch := &r.Archetypes[link.ArchId]
	newArch := &r.Archetypes[archId]

	newPos := newArch.Table.AddEntity(entityID)

	srcPage := oldArch.Table.GetChunk(link.Pos.ChunkIdx)
	dstPage := newArch.Table.GetChunk(newPos.ChunkIdx)

	newArch.Table.CopyColumnsFrom(&oldArch.Table, srcPage, link.Pos.ChunkSlot, dstPage, newPos.ChunkSlot)

	swappedEntity, swapped := oldArch.Table.SwapRemoveEntity(link.Pos)
	if swapped {
		r.EntityIndex.Upsert(swappedEntity, link.ArchId, link.Pos)
	}

	r.EntityIndex.Upsert(entityID, newArch.Id, newPos)
	return newPos
}

func (r *Catalog) Reset() {
	for i := int(RootID); i < int(r.lastArchetypeId); i++ {
		r.Archetypes[i].Reset()
	}
	clear(r.Archetypes[:])
	r.maskIndex.Reset()
	r.EntityIndex.Reset()
	r.lastArchetypeId = RootID
	r.addArchetype(comp.Composition{})
}
