package arch

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/internal/mem"
)

type ArchetypeRegistry struct {
	archetypeMaskMap   core.ArchetypeMaskMap
	Archetypes         [core.MaxArchetypeId]Archetype
	lastArchetypeId    core.ArchetypeId
	EntityLinkStore    EntityLinkStore
	componentsRegistry *core.ComponentsRegistry
	observer           ArchetypeObserver
	sharedColsInfos    []core.ComponentInfo
}

func NewArchetypeRegistry(
	componentsRegistry *core.ComponentsRegistry,
	observer ArchetypeObserver,
	initialEntityCap int,
) *ArchetypeRegistry {
	r := &ArchetypeRegistry{
		EntityLinkStore:    NewEntityLinkStore(initialEntityCap),
		componentsRegistry: componentsRegistry,
		observer:           observer,
		lastArchetypeId:    core.RootArchetypeId,
	}

	rootMask := core.ArchetypeMask{}
	r.InitArchetype(rootMask)

	return r
}

func (r *ArchetypeRegistry) InitArchetype(mask core.ArchetypeMask) core.ArchetypeId {
	if r.lastArchetypeId >= core.MaxArchetypeId {
		panic(fmt.Sprintf("Max archetype number exceeded: %d", core.MaxArchetypeId))
	}

	r.sharedColsInfos = r.sharedColsInfos[:0]

	defer func() {
		clear(r.sharedColsInfos)
		r.sharedColsInfos = r.sharedColsInfos[:0]
	}()

	for id := range mask.AllSet() {
		info := r.componentsRegistry.ByID(id)
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

func (r *ArchetypeRegistry) LastArchetypeId() core.ArchetypeId {
	return r.lastArchetypeId
}

func (r *ArchetypeRegistry) Get(mask core.ArchetypeMask) core.ArchetypeId {
	if archId, ok := r.archetypeMaskMap.Get(mask); ok {
		return archId
	}
	return core.NullArchetypeId
}

func (r *ArchetypeRegistry) AddEntity(entity uid.UID64, archId core.ArchetypeId) (mem.PageIdx, mem.PageSlot) {
	arch := &r.Archetypes[archId]
	pageIdx, pageSlot := arch.AddEntity(entity)
	r.EntityLinkStore.Update(entity, archId, pageIdx, pageSlot)
	return pageIdx, pageSlot
}

func (r *ArchetypeRegistry) UnlinkEntity(entity uid.UID64) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}

	linkArch := &r.Archetypes[link.ArchId]
	swappedEntity, swapped := linkArch.SwapRemoveEntity(link.PageIdx, link.PageSlot)

	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.ArchId, link.PageIdx, link.PageSlot)
	}

	r.EntityLinkStore.Clear(entity)
}

var ErrEntityNotFound = errors.New("entity not found in registry")

func (r *ArchetypeRegistry) AllocateComponentMemory(entity uid.UID64, compInfo core.ComponentInfo) (unsafe.Pointer, error) {
	compID := compInfo.ID

	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return nil, ErrEntityNotFound
	}

	currentArch := &r.Archetypes[link.ArchId]

	var targetArch *Archetype
	var targetPageIdx mem.PageIdx
	var targetRow mem.PageSlot

	if currentArch.Mask.IsSet(compID) {
		targetArch = currentArch
		targetPageIdx = link.PageIdx
		targetRow = link.PageSlot
	} else {
		nextArchID := r.ensureNextEdgeId(compID, currentArch)
		targetPageIdx, targetRow = r.moveEntity(entity, link, nextArchID)
		targetArch = &r.Archetypes[nextArchID]
	}

	if compInfo.Size == 0 {
		return nil, nil
	}

	column := targetArch.GetColumn(compID)
	page := targetArch.Memory.GetPage(targetPageIdx)
	return column.GetPointer(page, targetRow), nil
}

func (r *ArchetypeRegistry) ensureNextEdgeId(compID core.ComponentID, oldArch *Archetype) core.ArchetypeId {
	if nextEdgeId := oldArch.graph.edgesNext[compID]; nextEdgeId != core.NullArchetypeId {
		return nextEdgeId
	}

	newMask := oldArch.Mask.Set(compID)
	nextArchId := r.GetOrRegister(newMask)

	nextArch := &r.Archetypes[nextArchId]
	oldArch.linkNextArch(nextArch, compID)

	return nextArchId
}

func (r *ArchetypeRegistry) UnAssign(entity uid.UID64, compInfo core.ComponentInfo) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}

	oldArchId := link.ArchId
	compID := compInfo.ID
	oldArch := &r.Archetypes[oldArchId]

	if prevArchId := oldArch.graph.edgesPrev[compID]; prevArchId != core.NullArchetypeId {
		prevArch := &r.Archetypes[prevArchId]

		if prevArch.Mask.IsEmpty() {
			r.UnlinkEntity(entity)
			return
		}

		r.moveEntity(entity, link, prevArchId)
		return
	}

	newMask := oldArch.Mask.Clear(compID)

	if oldArch.Mask == newMask {
		return
	}

	if newMask.IsEmpty() {
		r.UnlinkEntity(entity)
		return
	}

	newPrevArchId := r.GetOrRegister(newMask)
	newPrevArch := &r.Archetypes[newPrevArchId]

	oldArch.linkPrevArch(newPrevArch, compID)

	r.moveEntity(entity, link, newPrevArchId)
}

func (r *ArchetypeRegistry) Reset() {
	for i := int(core.RootArchetypeId); i < int(r.lastArchetypeId); i++ {
		r.Archetypes[i].Reset()
	}
	clear(r.Archetypes[:])

	r.archetypeMaskMap.Reset()
	r.EntityLinkStore.Reset()

	clear(r.sharedColsInfos)
	r.sharedColsInfos = r.sharedColsInfos[:0]

	r.lastArchetypeId = core.RootArchetypeId
	rootMask := core.ArchetypeMask{}
	r.InitArchetype(rootMask)
}

func (r *ArchetypeRegistry) GetOrRegister(mask core.ArchetypeMask) core.ArchetypeId {
	if found, ok := r.archetypeMaskMap.Get(mask); ok {
		return found
	}
	archId := r.InitArchetype(mask)
	r.observer.OnArchetypeCreated(&r.Archetypes[archId])
	return archId
}

func (r *ArchetypeRegistry) moveEntity(entity uid.UID64, link EntityArchLink, archId core.ArchetypeId) (mem.PageIdx, mem.PageSlot) {
	oldArch := &r.Archetypes[link.ArchId]
	newArch := &r.Archetypes[archId]

	newPageIdx, newSlot := newArch.AddEntity(entity)

	srcPage := oldArch.Memory.GetPage(link.PageIdx)
	dstPage := newArch.Memory.GetPage(newPageIdx)

	for i := range newArch.Columns {
		dstCol := &newArch.Columns[i]

		if dstCol.CompID == core.EntityID {
			continue
		}

		if srcCol := oldArch.GetColumn(dstCol.CompID); srcCol != nil {
			srcPtr := srcCol.GetPointer(srcPage, link.PageSlot)
			dstPtr := dstCol.GetPointer(dstPage, newSlot)
			mem.CopyMemory(dstPtr, srcPtr, dstCol.ItemSize)
		}
	}

	swappedEntity, swapped := oldArch.SwapRemoveEntity(link.PageIdx, link.PageSlot)

	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.ArchId, link.PageIdx, link.PageSlot)
	}

	r.EntityLinkStore.Update(entity, newArch.Id, newPageIdx, newSlot)

	return newPageIdx, newSlot
}
