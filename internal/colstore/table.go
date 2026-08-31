package colstore

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/chunk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
)

// Pos is the storage position of an entity within a Table (chunk index + slot).
type Pos = chunk.Pos

// Idx is the index of a chunk within a Table.
type Idx = chunk.Idx

// Slot is the index of an entity slot within a chunk.
type Slot = chunk.Slot

// IDSeeder fills dst with valid entity IDs starting at (ptr, slot), including
// any bookkeeping such as index registration.
type IDSeeder func(dst []uid.UID64, ptr unsafe.Pointer, slot Slot)

// ColBake is a tracked column's pre-computed layout, for cursor filling.
type ColBake struct {
	Offset   uintptr
	CompSize uintptr
}

type Table struct {
	chunkPack  chunk.Pack
	columns    []ColDef
	compColIdx columnIndex
	seedIDs    IDSeeder

	// version counts structural changes that remove or relocate existing
	// entities (RemoveAt, compaction); appends never invalidate stored
	// entities, so they don't bump it. Captured in bulk.ChunkSnapshot.
	version uint32
}

func (t *Table) Version() uint32 { return t.version }

// --- Initialization ---

func (t *Table) SetIDSeeder(s IDSeeder) { t.seedIDs = s }

func (t *Table) Init(compDefs []comp.Def) {
	var layout chunk.Layout
	layout.Init(compDefs)
	t.chunkPack.Init(layout)

	count := len(compDefs) + 1
	t.columns = make([]ColDef, count)
	t.compColIdx.Reset()

	t.columns[entityColumnPos] = ColDef{
		CompID:   comp.EntityID,
		CompSize: unsafe.Sizeof(uid.UID64(0)),
		Offset:   t.chunkPack.Layout.Offsets[0],
	}
	for i, compDef := range compDefs {
		localIdx := columnPos(i + 1)
		t.compColIdx.Set(compDef.ID, localIdx)
		t.columns[localIdx] = ColDef{
			CompID:   compDef.ID,
			CompSize: compDef.Size,
			Offset:   t.chunkPack.Layout.Offsets[i+1],
		}
	}
}

// --- Bake ---

// BakeColumns pre-computes offset and size for each tracked column.
func (t *Table) BakeColumns(defs []comp.Def) []ColBake {
	result := make([]ColBake, len(defs))
	for i, def := range defs {
		if col := t.getColumn(def.ID); col != nil {
			result[i] = ColBake{Offset: col.Offset, CompSize: col.CompSize}
		}
	}
	return result
}

func (t *Table) BakeOffsets(ids []comp.ID) []uintptr {
	offsets := make([]uintptr, len(ids))
	for i, id := range ids {
		if col := t.getColumn(id); col != nil {
			offsets[i] = col.Offset
		}
	}
	return offsets
}

// BakeOptional is BakeOffsets plus a per-id presence flag — present via a
// data column when the id has one, or via mask membership when it's a tag
// (which has no column at all).
func (t *Table) BakeOptional(ids []comp.ID, mask comp.Mask) (offsets []uintptr, present []bool) {
	offsets = make([]uintptr, len(ids))
	present = make([]bool, len(ids))
	for i, id := range ids {
		if col := t.getColumn(id); col != nil {
			offsets[i] = col.Offset
			present[i] = true
		} else if mask.IsSet(id) {
			present[i] = true // tag: present, no column, offset stays 0 (never dereferenced — size 0)
		}
	}
	return offsets, present
}

// --- Read ---

func (t *Table) Len() uint32 { return t.chunkPack.Len() }

// ComponentAt returns a pointer to component id at (ptr, slot); nil if untracked.
func (t *Table) ComponentAt(ptr unsafe.Pointer, slot Slot, id comp.ID) unsafe.Pointer {
	col := t.getColumn(id)
	if col == nil {
		return nil
	}
	return col.At(ptr, slot)
}

func (t *Table) ChunkPtrAt(idx Idx) unsafe.Pointer {
	return t.chunkPack.ChunkPtr(idx)
}

func (t *Table) ChunkIdxByPtr(ptr unsafe.Pointer) Idx {
	return t.chunkPack.ChunkIdxByPtr(ptr)
}

// FillCursorNext fills cur with the next non-empty chunk at or after from;
// returns its index, or false when no chunks remain.
func (t *Table) FillCursorNext(cur *iter.Cursor, from int, offsets []uintptr) (int, bool) {
	idx, _, _, ok := t.chunkPack.NextNonEmptyChunk(from)
	if !ok {
		return from, false
	}
	ptr := t.chunkPack.ChunkPtr(chunk.Idx(idx))
	cur.Base = ptr
	cur.Offsets = offsets
	cur.IDs = unsafe.Slice((*uid.UID64)(t.columns[entityColumnPos].At(ptr, 0)), int(t.chunkPack.ChunkLen(chunk.Idx(idx))))
	return idx, true
}

// --- Write ---

// AllocSlots extends chunk idx by n slots without seeding entity IDs;
// returns the chunk base pointer and the starting slot of the new range.
func (t *Table) AllocSlots(idx Idx, n int) (unsafe.Pointer, Slot) {
	return t.chunkPack.Extend(idx, n)
}

func (t *Table) SetEntityRange(ptr unsafe.Pointer, slot Slot, ids []uid.UID64) {
	dst := unsafe.Slice((*uid.UID64)(t.columns[entityColumnPos].At(ptr, slot)), len(ids))
	copy(dst, ids)
}

// CopyRangeFrom copies n consecutive slots' component columns from src,
// one CopyMemory per matched column; components absent in src are skipped.
// The entity ID column is not touched — call SetEntityRange separately.
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

// MoveEntityFrom moves entityID from src into a freshly allocated slot here,
// copying matching columns, then swap-removes the source slot. Returns the new
// position plus the entity displaced by the swap, if any.
func (t *Table) MoveEntityFrom(src *Table, entityID uid.UID64, srcPtr unsafe.Pointer, srcSlot Slot) (newPtr unsafe.Pointer, newSlot Slot, swappedEntity uid.UID64, swapped bool) {
	newPos := t.chunkPack.AllocSlot()
	newPtr = t.chunkPack.ChunkPtr(newPos.Idx)
	newSlot = newPos.Slot
	*(*uid.UID64)(t.columns[entityColumnPos].At(newPtr, newSlot)) = entityID

	for i := firstDataColumnPos; int(i) < len(t.columns); i++ {
		dstCol := &t.columns[i]
		srcCol := src.getColumn(dstCol.CompID)
		if srcCol == nil {
			continue
		}
		chunk.CopyMemory(
			dstCol.At(newPtr, newSlot),
			srcCol.At(srcPtr, srcSlot),
			dstCol.CompSize,
		)
	}

	swappedEntity, swapped = src.RemoveAt(srcPtr, srcSlot)
	return
}

// SpawnCursor allocates and seeds n entity slots in chunk idx, filling cur
// with offsets adjusted for the starting slot.
func (t *Table) SpawnCursor(cur *iter.Cursor, idx Idx, n int, colBakes []ColBake) ([]uid.UID64, Pos) {
	base, slot, ids := t.spawnEntitySlice(idx, n)
	cur.Base = base
	cur.IDs = ids
	for i, colBake := range colBakes {
		cur.Offsets[i] = colBake.Offset + uintptr(slot)*colBake.CompSize
	}
	return ids, Pos{Idx: idx, Slot: slot}
}

// ReserveSlots ensures capacity for count entities; returns the first chunk
// index, slots available there, and the capacity of each subsequent chunk.
func (t *Table) ReserveSlots(count int) (firstIdx Idx, firstAvailable int, chunkCap int) {
	idx, avail := t.chunkPack.ReserveSlots(count)
	return idx, avail, int(t.chunkPack.Layout.ChunkCap)
}

// ReleaseSlots clears the Reserved marker set by ReserveSlots.
func (t *Table) ReleaseSlots() {
	t.chunkPack.Reserved = 0
}

// Purge releases all trailing empty chunks, including the spare kept for
// reuse — for tables not expected to repopulate soon.
func (t *Table) Purge() {
	t.chunkPack.Purge()
}

// RemoveAt swap-and-pops the slot at (ptr, slot); returns the ID that moved
// into it, or (0, false) if it was already the tail.
func (t *Table) RemoveAt(ptr unsafe.Pointer, slot Slot) (uid.UID64, bool) {
	t.version++
	lastChunkIdx, lastSlot := t.chunkPack.ResolveTail()
	lastPtr := t.chunkPack.ChunkPtr(lastChunkIdx)

	if ptr == lastPtr && slot == lastSlot {
		t.zeroSlot(lastPtr, lastSlot)
		t.chunkPack.FreeSlot(lastChunkIdx)
		return 0, false
	}

	entityToMove := *(*uid.UID64)(t.columns[entityColumnPos].At(lastPtr, lastSlot))
	t.swapCopy(ptr, slot, lastPtr, lastSlot)
	t.zeroSlot(lastPtr, lastSlot)
	t.chunkPack.FreeSlot(lastChunkIdx)

	return entityToMove, true
}

func (t *Table) Clear() {
	t.version++
	t.chunkPack.Clear()
	t.columns = nil
	t.compColIdx.Reset()
}

// --- Internal ---

func (t *Table) getColumn(id comp.ID) *ColDef {
	localIdx := t.compColIdx.Get(id)
	if localIdx == invalidColumnPos {
		return nil
	}
	return &t.columns[localIdx]
}

func (t *Table) spawnEntitySlice(idx chunk.Idx, n int) (base unsafe.Pointer, slot chunk.Slot, ids []uid.UID64) {
	base, slot = t.chunkPack.Extend(idx, n)
	ids = unsafe.Slice((*uid.UID64)(t.columns[entityColumnPos].At(base, slot)), n)
	t.seedIDs(ids, base, slot)
	return
}

func (t *Table) zeroSlot(chunkPtr unsafe.Pointer, slot chunk.Slot) {
	for i := range t.columns {
		col := &t.columns[i]
		chunk.ZeroMemory(col.At(chunkPtr, slot), col.CompSize)
	}
}

// zeroRange zeroes n consecutive slots starting at slot — one ZeroMemory per column.
func (t *Table) zeroRange(chunkPtr unsafe.Pointer, slot chunk.Slot, n int) {
	for i := range t.columns {
		col := &t.columns[i]
		chunk.ZeroMemory(col.At(chunkPtr, slot), uintptr(n)*col.CompSize)
	}
}

// moveRange copies n slots from srcSlot to dstSlot within one chunk, entity
// column included; ranges may overlap (memmove semantics).
func (t *Table) moveRange(chunkPtr unsafe.Pointer, dstSlot, srcSlot chunk.Slot, n int) {
	for i := range t.columns {
		col := &t.columns[i]
		chunk.CopyMemory(
			col.At(chunkPtr, dstSlot),
			col.At(chunkPtr, srcSlot),
			uintptr(n)*col.CompSize,
		)
	}
}

func (t *Table) swapCopy(dstPtr unsafe.Pointer, dstSlot chunk.Slot, srcPtr unsafe.Pointer, srcSlot chunk.Slot) {
	for i := range t.columns {
		col := &t.columns[i]
		chunk.CopyMemory(
			col.At(dstPtr, dstSlot),
			col.At(srcPtr, srcSlot),
			col.CompSize,
		)
	}
}
