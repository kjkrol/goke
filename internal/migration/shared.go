package migration

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/addr"
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/colstore"
)

// resolveSlotRefs resolves ids to their current SlotRef positions within
// srcTable, using *idsScratch/*slotRefsScratch as the caller's persistent
// scratch storage (grown lazily, never shrunk). While snap.TableVer still
// matches srcTable's structural version, every id is alive at its captured
// position and ids is returned unchanged (no copy) — SlotAligned synthesizes
// the slots outright, otherwise one unchecked Slot read per id suffices. On
// a version mismatch, every id is revalidated against addrBook; dead or
// relocated-away ids are skipped and *idsScratch backs the returned
// (possibly shorter) id slice.
func resolveSlotRefs(
	addrBook *addr.Book,
	srcTable *colstore.Table,
	snap bulk.ChunkSnapshot,
	ids []uid.UID64,
	idsScratch *[]uid.UID64,
	slotRefsScratch *[]colstore.SlotRef,
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
	var lastPtr unsafe.Pointer
	var lastIdx colstore.Idx
	valid := 0
	for _, id := range ids {
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
		valid++
	}
	return validIDsBuf[:valid], refs[:valid]
}

// removeBatch removes ids outright from srcTable: one batched Compact call
// closes every vacated hole at once, then the address book is updated in
// one homogeneous pass — deletions (id recycling) first, survivor
// relocations from the compaction's SlotMove list second. Used both by
// Remover.Migrate and by Migrator.applyGroup when a batch's destination
// composition resolves empty (every component removed).
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
