package colstore

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/mem"
)

// IDSeeder fills dst with valid entity IDs and performs any associated bookkeeping
// (e.g. index registration). It is set once per archetype via SetIDSeeder.
type IDSeeder func(dst []uid.UID64, chunkIdx mem.ChunkIdx, startSlot mem.ChunkSlot)

type Table struct {
	mem.Block
	columns    []Column
	compColIdx columnIndex
	seedIDs    IDSeeder
}

func (t *Table) SetIDSeeder(s IDSeeder) { t.seedIDs = s }

func (t *Table) Init(compMetas []comp.Meta) {
	var layout mem.ChunkLayout
	layout.Init(compMetas)
	t.Block.Init(layout)

	count := len(compMetas) + 1
	t.columns = make([]Column, count)
	t.compColIdx.Reset()

	t.columns[entityColumnPos] = Column{
		CompID:   comp.EntityID,
		CompSize: unsafe.Sizeof(uid.UID64(0)),
		Offset:   t.Layout.Offsets[0],
	}
	for i, compMeta := range compMetas {
		localIdx := columnPos(i + 1)
		t.compColIdx.Set(compMeta.ID, localIdx)
		t.columns[localIdx] = Column{
			CompID:   compMeta.ID,
			CompSize: compMeta.Size,
			Offset:   t.Layout.Offsets[i+1],
		}
	}
}

func (t *Table) NumColumns() int {
	return len(t.columns)
}

func (t *Table) GetEntityColumn() *Column {
	return &t.columns[entityColumnPos]
}

func (t *Table) GetColumn(id comp.ID) *Column {
	localIdx := t.compColIdx.Get(id)
	if localIdx == invalidColumnPos {
		return nil
	}
	return &t.columns[localIdx]
}

func (t *Table) CopyColumnsFrom(src *Table, srcPos mem.BlockPos, dstPos mem.BlockPos) {
	srcPtr := src.ChunkPtr(srcPos.ChunkIdx)
	dstPtr := t.ChunkPtr(dstPos.ChunkIdx)
	for i := firstDataColumnPos; int(i) < len(t.columns); i++ {
		dstCol := &t.columns[i]
		srcCol := src.GetColumn(dstCol.CompID)
		if srcCol == nil {
			continue
		}
		mem.CopyMemory(
			dstCol.At(dstPtr, dstPos.ChunkSlot),
			srcCol.At(srcPtr, srcPos.ChunkSlot),
			dstCol.CompSize,
		)
	}
}

func (t *Table) SpawnEntitySlice(idx mem.ChunkIdx, n int) (base unsafe.Pointer, slot mem.ChunkSlot, ids []uid.UID64) {
	slot = t.ChunkLen(idx)
	t.AllocSlots(idx, n)
	base = t.ChunkPtr(idx)
	ids = unsafe.Slice((*uid.UID64)(t.GetEntityColumn().At(base, slot)), n)
	t.seedIDs(ids, idx, slot)
	return
}

func (t *Table) AddEntity(entityID uid.UID64) mem.BlockPos {
	pos := t.AllocSlot()
	*(*uid.UID64)(t.GetEntityColumn().At(t.ChunkPtr(pos.ChunkIdx), pos.ChunkSlot)) = entityID
	return pos
}

func (t *Table) SwapRemoveEntity(pos mem.BlockPos) (uid.UID64, bool) {
	lastChunkIdx, lastSlot := t.ResolveTail()
	lastPtr := t.ChunkPtr(lastChunkIdx)

	if pos.ChunkIdx == lastChunkIdx && pos.ChunkSlot == lastSlot {
		t.zeroSlot(lastPtr, lastSlot)
		t.FreeSlot(lastChunkIdx)
		return 0, false
	}

	entityToMove := *(*uid.UID64)(t.GetEntityColumn().At(lastPtr, lastSlot))
	t.swapCopy(t.ChunkPtr(pos.ChunkIdx), pos.ChunkSlot, lastPtr, lastSlot)
	t.zeroSlot(lastPtr, lastSlot)
	t.FreeSlot(lastChunkIdx)

	return entityToMove, true
}

func (t *Table) zeroSlot(chunkPtr unsafe.Pointer, slot mem.ChunkSlot) {
	for i := range t.columns {
		col := &t.columns[i]
		mem.ZeroMemory(col.At(chunkPtr, slot), col.CompSize)
	}
}

func (t *Table) swapCopy(dstPtr unsafe.Pointer, dstSlot mem.ChunkSlot, srcPtr unsafe.Pointer, srcSlot mem.ChunkSlot) {
	for i := range t.columns {
		col := &t.columns[i]
		mem.CopyMemory(
			col.At(dstPtr, dstSlot),
			col.At(srcPtr, srcSlot),
			col.CompSize,
		)
	}
}

func (t *Table) BakeOffsets(metas []comp.Meta) []uintptr {
	offsets := make([]uintptr, len(metas))
	for i, meta := range metas {
		if col := t.GetColumn(meta.ID); col != nil {
			offsets[i] = col.Offset
		}
	}
	return offsets
}

func (t *Table) Clear() {
	t.Block.Clear()
	t.columns = nil
	t.compColIdx.Reset()
}
