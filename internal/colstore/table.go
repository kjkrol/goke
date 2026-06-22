package colstore

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/chunk"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

// Pos is the storage position of an entity within a Table (chunk index + slot).
type Pos = chunk.Pos

// Idx is the index of a chunk within a Table.
type Idx = chunk.Idx

// Slot is the index of an entity slot within a chunk.
type Slot = chunk.Slot

// IDSeeder fills dst with valid entity IDs starting at pos and performs any
// associated bookkeeping (e.g. index registration). It is set once per
// archetype via SetIDSeeder.
type IDSeeder func(dst []uid.UID64, pos Pos)

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
// at pos, or nil if the component is not tracked by this table.
func (t *Table) ComponentAt(pos Pos, id comp.ID) unsafe.Pointer {
	col := t.getColumn(id)
	if col == nil {
		return nil
	}
	return col.At(t.chunkPack.ChunkPtr(pos.Idx), pos.Slot)
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

// PointCursor positions cur at pos (chunk base + slot) without touching its
// column offsets. The caller sets cur.Offsets once per table, then moves the
// cursor across slots of that table with PointCursor.
func (t *Table) PointCursor(cur *iter.Cursor, pos Pos) {
	cur.Base = t.chunkPack.ChunkPtr(pos.Idx)
	cur.Slot = uintptr(pos.Slot)
}

// --- Write ---

// MoveEntityFrom moves entityID from src at srcPos into this table.
// It allocates a slot, writes the entity ID, copies matching component columns
// from src, then swap-removes the source slot.
// Returns the new position, the entity displaced by the swap, and whether a swap occurred.
func (dst *Table) MoveEntityFrom(src *Table, entityID uid.UID64, srcPos Pos) (newPos Pos, swappedEntity uid.UID64, swapped bool) {
	newPos = dst.chunkPack.AllocSlot()
	dstPtr := dst.chunkPack.ChunkPtr(newPos.Idx)
	*(*uid.UID64)(dst.columns[entityColumnPos].At(dstPtr, newPos.Slot)) = entityID

	srcPtr := src.chunkPack.ChunkPtr(srcPos.Idx)
	for i := firstDataColumnPos; int(i) < len(dst.columns); i++ {
		dstCol := &dst.columns[i]
		srcCol := src.getColumn(dstCol.CompID)
		if srcCol == nil {
			continue
		}
		chunk.CopyMemory(
			dstCol.At(dstPtr, newPos.Slot),
			srcCol.At(srcPtr, srcPos.Slot),
			dstCol.CompSize,
		)
	}

	swappedEntity, swapped = src.RemoveAt(srcPos)
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

// RemoveAt removes the slot at pos using swap-and-pop to keep the table dense.
// Returns the ID that moved into pos and true if a swap occurred,
// or (0, false) if pos was already the last slot.
func (t *Table) RemoveAt(pos Pos) (uid.UID64, bool) {
	lastChunkIdx, lastSlot := t.chunkPack.ResolveTail()
	lastPtr := t.chunkPack.ChunkPtr(lastChunkIdx)

	if pos.Idx == lastChunkIdx && pos.Slot == lastSlot {
		t.zeroSlot(lastPtr, lastSlot)
		t.chunkPack.FreeSlot(lastChunkIdx)
		return 0, false
	}

	entityToMove := *(*uid.UID64)(t.columns[entityColumnPos].At(lastPtr, lastSlot))
	t.swapCopy(t.chunkPack.ChunkPtr(pos.Idx), pos.Slot, lastPtr, lastSlot)
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
	t.seedIDs(ids, Pos{Idx: idx, Slot: slot})
	return
}

func (t *Table) zeroSlot(chunkPtr unsafe.Pointer, slot chunk.Slot) {
	for i := range t.columns {
		col := &t.columns[i]
		chunk.ZeroMemory(col.At(chunkPtr, slot), col.CompSize)
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
