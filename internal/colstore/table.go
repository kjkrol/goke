package colstore

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/soa"
)

type Table struct {
	soa.Block
	columns    []Column
	compColIdx columnIndex
}

func (t *Table) Init(compMetas []comp.Meta) {
	var layout soa.ChunkLayout
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

func (t *Table) ZeroSlot(chunk *soa.Chunk, slot soa.ChunkSlot) {
	for i := range t.columns {
		col := &t.columns[i]
		soa.ZeroMemory(col.At(chunk, slot), col.CompSize)
	}
}

func (t *Table) SwapCopy(dst *soa.Chunk, dstSlot soa.ChunkSlot, src *soa.Chunk, srcSlot soa.ChunkSlot) {
	for i := range t.columns {
		col := &t.columns[i]
		soa.CopyMemory(
			col.At(dst, dstSlot),
			col.At(src, srcSlot),
			col.CompSize,
		)
	}
}

func (t *Table) CopyColumnsFrom(src *Table, srcChunk *soa.Chunk, srcSlot soa.ChunkSlot, dstChunk *soa.Chunk, dstSlot soa.ChunkSlot) {
	for i := firstDataColumnPos; int(i) < len(t.columns); i++ {
		dstCol := &t.columns[i]
		srcCol := src.GetColumn(dstCol.CompID)
		if srcCol == nil {
			continue
		}
		soa.CopyMemory(
			dstCol.At(dstChunk, dstSlot),
			srcCol.At(srcChunk, srcSlot),
			dstCol.CompSize,
		)
	}
}

func (t *Table) AddEntity(entityID uid.UID64) soa.BlockPos {
	chunk, pos := t.AllocSlot()
	entityCol := t.GetEntityColumn()
	*(*uid.UID64)(entityCol.At(chunk, pos.ChunkSlot)) = entityID
	return pos
}

func (t *Table) SwapRemoveEntity(pos soa.BlockPos) (uid.UID64, bool) {
	lastChunkIdx, lastChunk := t.ResolveTail()
	lastSlot := soa.ChunkSlot(lastChunk.Len - 1)
	targetChunk := &t.Chunks[pos.ChunkIdx]

	if pos.ChunkIdx == lastChunkIdx && pos.ChunkSlot == lastSlot {
		t.ZeroSlot(lastChunk, lastSlot)
		lastChunk.Len--
		t.Len--
		return 0, false
	}

	entityToMove := *(*uid.UID64)(t.GetEntityColumn().At(lastChunk, lastSlot))
	t.SwapCopy(targetChunk, pos.ChunkSlot, lastChunk, lastSlot)
	t.ZeroSlot(lastChunk, lastSlot)
	lastChunk.Len--
	t.Len--

	return entityToMove, true
}

func (t *Table) Clear() {
	t.Block.Clear()
	t.columns = nil
	t.compColIdx.Reset()
}
