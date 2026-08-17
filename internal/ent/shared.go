package ent

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/colstore"
	"github.com/kjkrol/goke/v3/internal/comp"
)

// resolveSlotRefs resolves ids to their current SlotRef positions in
// srcTable, revalidating against addrBook (and dropping dead/relocated ids)
// only if snap is stale. origIdxScratch is optional; when non-nil it
// receives each survivor's original index into ids, for compacting a
// parallel per-id buffer (e.g. ValueEditor's payload) in lockstep.
func resolveSlotRefs(
	addrBook *addr.Book,
	srcTable *colstore.Table,
	snap bulk.ChunkSnapshot,
	ids []uid.UID64,
	idsScratch *[]uid.UID64,
	slotRefsScratch *[]colstore.SlotRef,
	origIdxScratch *[]int,
) (validIDs []uid.UID64, slotRefs []colstore.SlotRef) {
	n := len(ids)
	*slotRefsScratch = growTo(*slotRefsScratch, n)
	refs := *slotRefsScratch

	if snap.TableVer == srcTable.Version() {
		if snap.SlotAligned {
			for i := range n {
				refs[i] = colstore.SlotRef{Ptr: snap.ChunkPtr, Idx: snap.ChunkIdx, Slot: colstore.Slot(i)}
			}
		} else {
			for i, id := range ids {
				refs[i] = colstore.SlotRef{
					Ptr:  snap.ChunkPtr,
					Idx:  snap.ChunkIdx,
					Slot: addrBook.Index.GetUnchecked(id).Slot,
				}
			}
		}
		return ids, refs[:n]
	}

	// Version mismatch: something mutated the table since capture.
	*idsScratch = growTo(*idsScratch, n)
	validIDsBuf := *idsScratch
	var origIdx []int
	if origIdxScratch != nil {
		*origIdxScratch = growTo(*origIdxScratch, n)
		origIdx = *origIdxScratch
	}
	var lastPtr unsafe.Pointer
	var lastIdx colstore.Idx
	valid := 0
	for i, id := range ids {
		entry, ok := addrBook.Get(id)
		if !ok || entry.ArchID != snap.ArchID {
			continue
		}
		if entry.ChunkPtr != lastPtr {
			lastPtr = entry.ChunkPtr
			lastIdx = srcTable.ChunkIdxByPtr(lastPtr)
		}
		validIDsBuf[valid] = id
		refs[valid] = colstore.SlotRef{Ptr: entry.ChunkPtr, Idx: lastIdx, Slot: entry.Slot}
		if origIdx != nil {
			origIdx[valid] = i
		}
		valid++
	}
	return validIDsBuf[:valid], refs[:valid]
}

// resolveDst computes the destination archetype for applying spec's adds
// then dels to srcArchID's composition. Returns arch.NullID when nothing is
// left (every component removed).
func resolveDst(archCatalog *arch.Catalog, spec comp.EditSpec, srcArchID arch.ID) arch.ID {
	set := archCatalog.Archetypes[srcArchID].Composition()
	for i := range spec.AddDefs {
		d := spec.AddDefs[i]
		if !set.Mask.IsSet(d.ID) {
			set = set.With(d)
		}
	}
	for i := range spec.DelDefs {
		set = set.Without(spec.DelDefs[i].ID)
	}
	if set.Mask.IsEmpty() {
		return arch.NullID
	}
	return archCatalog.Upsert(set)
}

// removeBatch removes ids outright from srcTable, compacting and updating
// addrBook in one pass. Used by Remover.Migrate and by Editor.applyGroup
// when a batch's destination composition resolves empty.
func removeBatch(
	addrBook *addr.Book,
	defrag *colstore.Defragmenter,
	srcArchID arch.ID,
	srcTable *colstore.Table,
	ids []uid.UID64,
	slotRefs []colstore.SlotRef,
) {
	moves := defrag.Compact(srcTable, slotRefs)
	for _, id := range ids {
		addrBook.Delete(id)
	}
	for _, sm := range moves {
		addrBook.MoveUnchecked(sm.ID, srcArchID, sm.NewPtr, sm.NewSlot)
	}
}

// growTo returns s resized to n elements, reallocating only when needed.
func growTo[T any](s []T, n int) []T {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]T, n)
}
