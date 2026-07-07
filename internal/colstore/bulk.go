package colstore

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/chunk"
)

// SlotRef identifies an entity slot by its chunk memory pointer and slot index.
// Unlike Pos, SlotRef remains valid across SwapChunks operations because it
// addresses the chunk's backing memory directly, not its position in the Pack.
// Idx is the chunk's position in the Pack at the time of compaction; it is
// filled by Defragmenter.Compact and must not be set by callers.
type SlotRef struct {
	Ptr  unsafe.Pointer
	Idx  Idx
	Slot Slot
}

// SlotMove records a surviving entity that was relocated within a table
// during CompactHoles. Used by the migration layer for batch addr.Book updates.
type SlotMove struct {
	ID      uid.UID64
	NewPtr  unsafe.Pointer
	NewSlot Slot
}

// AllocSlots extends chunk idx by n slots without seeding entity IDs.
// Returns the chunk base pointer and the starting slot of the new range.
// Used by the migration layer to claim space in a target archetype before
// copying component data from the source.
func (t *Table) AllocSlots(idx Idx, n int) (unsafe.Pointer, Slot) {
	return t.chunkPack.Extend(idx, n)
}

// SetEntityRange writes ids into the entity ID column starting at (ptr, slot).
func (t *Table) SetEntityRange(ptr unsafe.Pointer, slot Slot, ids []uid.UID64) {
	dst := unsafe.Slice((*uid.UID64)(t.columns[entityColumnPos].At(ptr, slot)), len(ids))
	copy(dst, ids)
}

// CopyRangeFrom copies n consecutive slots' component columns from src starting
// at (srcPtr, srcSlot) into this table starting at (dstPtr, dstSlot), matching
// by component ID — one CopyMemory per matched column, not per slot. Components
// absent in src are skipped. The entity ID column is not touched; call
// SetEntityRange separately.
func (t *Table) CopyRangeFrom(src *Table, srcPtr unsafe.Pointer, srcSlot Slot, dstPtr unsafe.Pointer, dstSlot Slot, n int) {
	for i := firstDataColumnPos; int(i) < len(t.columns); i++ {
		dstCol := &t.columns[i]
		srcCol := src.getColumn(dstCol.CompID)
		if srcCol == nil {
			continue
		}
		chunk.CopyMemory(
			dstCol.At(dstPtr, dstSlot),
			srcCol.At(srcPtr, srcSlot),
			uintptr(n)*dstCol.CompSize,
		)
	}
}
