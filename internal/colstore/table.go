package colstore

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/chunk"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/iter"
)

// Pos is the storage position of an entity within a Table (chunk index + slot).
type Pos = chunk.Pos

// Idx is the index of a chunk within a Table.
type Idx = chunk.Idx

// Slot is the index of an entity slot within a chunk.
type Slot = chunk.Slot

// IDSeeder fills dst with valid entity IDs starting at (ptr, slot) and performs
// any associated bookkeeping (e.g. index registration). It is set once per
// archetype via SetIDSeeder.
type IDSeeder func(dst []uid.UID64, ptr unsafe.Pointer, slot Slot)

// ColBake holds the layout info needed to fill a cursor for a tracked column.
// Obtained via BakeColumns; passed to SpawnCursor.
type ColBake struct {
	Offset   uintptr
	CompSize uintptr
}

type Table struct {
	chunkPack  chunk.Pack
	columns    []ColDef
	compColIdx columnIndex
	seedIDs    IDSeeder
}

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

// BakeColumns returns a ColBake slice for the given component defs,
// pre-computing offset and size for each tracked column.
func (t *Table) BakeColumns(defs []comp.Def) []ColBake {
	result := make([]ColBake, len(defs))
	for i, def := range defs {
		if col := t.getColumn(def.ID); col != nil {
			result[i] = ColBake{Offset: col.Offset, CompSize: col.CompSize}
		}
	}
	return result
}

// BakeOffsets returns a uintptr slice with the base offset of each tracked column.
func (t *Table) BakeOffsets(ids []comp.ID) []uintptr {
	offsets := make([]uintptr, len(ids))
	for i, id := range ids {
		if col := t.getColumn(id); col != nil {
			offsets[i] = col.Offset
		}
	}
	return offsets
}

// --- Read ---

func (t *Table) Len() uint32 { return t.chunkPack.Len() }

// ComponentAt returns an unsafe.Pointer to the component with the given id
// at (ptr, slot), or nil if the component is not tracked by this table.
func (t *Table) ComponentAt(ptr unsafe.Pointer, slot Slot, id comp.ID) unsafe.Pointer {
	col := t.getColumn(id)
	if col == nil {
		return nil
	}
	return col.At(ptr, slot)
}

// ChunkPtrAt returns the backing memory pointer for chunk at idx.
func (t *Table) ChunkPtrAt(idx Idx) unsafe.Pointer {
	return t.chunkPack.ChunkPtr(idx)
}

// ChunkIdxByPtr returns the chunk index whose backing pointer equals ptr.
func (t *Table) ChunkIdxByPtr(ptr unsafe.Pointer) Idx {
	return t.chunkPack.ChunkIdxByPtr(ptr)
}

// FillCursorNext advances to the next non-empty chunk starting at from,
// filling cur with its base pointer, offsets, and entity ID slice.
// Returns the chunk index and false when no more chunks remain.
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

// MoveEntityFrom moves entityID from src at (srcPtr, srcSlot) into this table.
// It allocates a slot, writes the entity ID, copies matching component columns
// from src, then swap-removes the source slot.
// Returns the new chunk pointer, new slot, the entity displaced by the swap,
// and whether a swap occurred.
func (dst *Table) MoveEntityFrom(src *Table, entityID uid.UID64, srcPtr unsafe.Pointer, srcSlot Slot) (newPtr unsafe.Pointer, newSlot Slot, swappedEntity uid.UID64, swapped bool) {
	newPos := dst.chunkPack.AllocSlot()
	newPtr = dst.chunkPack.ChunkPtr(newPos.Idx)
	newSlot = newPos.Slot
	*(*uid.UID64)(dst.columns[entityColumnPos].At(newPtr, newSlot)) = entityID

	for i := firstDataColumnPos; int(i) < len(dst.columns); i++ {
		dstCol := &dst.columns[i]
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

// SpawnCursor allocates n entity slots starting at idx, invokes the IDSeeder,
// and fills cur with the base pointer and per-column offsets adjusted for the
// starting slot.
func (t *Table) SpawnCursor(cur *iter.Cursor, idx Idx, n int, colBakes []ColBake) ([]uid.UID64, Pos) {
	base, slot, ids := t.spawnEntitySlice(idx, n)
	cur.Base = base
	cur.IDs = ids
	for i, colBake := range colBakes {
		cur.Offsets[i] = colBake.Offset + uintptr(slot)*colBake.CompSize
	}
	return ids, Pos{Idx: idx, Slot: slot}
}

// ReserveSlots ensures enough chunks exist for count entities and returns the first
// chunk index, the number of slots available in that first chunk, and the
// capacity of each subsequent chunk.
func (t *Table) ReserveSlots(count int) (firstIdx Idx, firstAvailable int, chunkCap int) {
	idx, avail := t.chunkPack.ReserveSlots(count)
	return idx, avail, int(t.chunkPack.Layout.ChunkCap)
}

// ReleaseSlots clears the Reserved marker set by ReserveSlots.
func (t *Table) ReleaseSlots() {
	t.chunkPack.Reserved = 0
}

// Purge releases all trailing empty chunks, including the one spare chunk
// normally kept for reuse after a table drains to zero rows. Call it when
// the table is not expected to be repopulated soon, to give its memory back
// immediately instead of waiting for natural reuse.
func (t *Table) Purge() {
	t.chunkPack.Purge()
}

// RemoveAt removes the slot at (ptr, slot) using swap-and-pop to keep the
// table dense. Returns the ID that moved into (ptr, slot) and true if a swap
// occurred, or (0, false) if that slot was already the last one.
func (t *Table) RemoveAt(ptr unsafe.Pointer, slot Slot) (uid.UID64, bool) {
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

// moveRange copies n consecutive slots from srcSlot to dstSlot within the chunk
// at chunkPtr — one CopyMemory per column, entity ID column included. The ranges
// may overlap; CopyMemory has memmove semantics.
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
