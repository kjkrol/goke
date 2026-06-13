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

func (cs *Table) Init(compMetas []comp.Meta) {
	layout := soa.CalculateLayout(compMetas)
	cs.Block = soa.NewBlock(layout)

	count := len(compMetas) + 1
	cs.columns = make([]Column, count)
	cs.compColIdx.Reset()

	cs.columns[entityColumnPos] = Column{
		CompID:     comp.EntityID,
		ItemSize:   unsafe.Sizeof(uid.UID64(0)),
		PageOffset: cs.Layout.Offsets[0],
	}
	for i, compMeta := range compMetas {
		localIdx := columnPos(i + 1)
		cs.compColIdx.Set(compMeta.ID, localIdx)
		cs.columns[localIdx] = Column{
			CompID:     compMeta.ID,
			ItemSize:   compMeta.Size,
			PageOffset: cs.Layout.Offsets[i+1],
		}
	}
}

func (cs *Table) NumColumns() int {
	return len(cs.columns)
}

func (cs *Table) GetEntityColumn() *Column {
	return &cs.columns[entityColumnPos]
}

func (cs *Table) GetColumn(id comp.ID) *Column {
	localIdx := cs.compColIdx.Get(id)
	if localIdx == invalidColumnPos {
		return nil
	}
	return &cs.columns[localIdx]
}

func (cs *Table) ZeroSlot(chunk *soa.Chunk, slot soa.ChunkSlot) {
	for i := range cs.columns {
		col := &cs.columns[i]
		soa.ZeroMemory(col.At(chunk, slot), col.ItemSize)
	}
}

func (cs *Table) SwapCopy(dst *soa.Chunk, dstSlot soa.ChunkSlot, src *soa.Chunk, srcSlot soa.ChunkSlot) {
	for i := range cs.columns {
		col := &cs.columns[i]
		soa.CopyMemory(
			col.At(dst, dstSlot),
			col.At(src, srcSlot),
			col.ItemSize,
		)
	}
}

func (cs *Table) CopyColumnsFrom(src *Table, srcChunk *soa.Chunk, srcSlot soa.ChunkSlot, dstChunk *soa.Chunk, dstSlot soa.ChunkSlot) {
	for i := firstDataColumnPos; int(i) < len(cs.columns); i++ {
		dstCol := &cs.columns[i]
		srcCol := src.GetColumn(dstCol.CompID)
		if srcCol == nil {
			continue
		}
		soa.CopyMemory(
			dstCol.At(dstChunk, dstSlot),
			srcCol.At(srcChunk, srcSlot),
			dstCol.ItemSize,
		)
	}
}

func (cs *Table) AddEntity(entityID uid.UID64) soa.BlockPos {
	chunk, pos := cs.AllocSlot()
	entityCol := cs.GetEntityColumn()
	*(*uid.UID64)(entityCol.At(chunk, pos.ChunkSlot)) = entityID
	return pos
}

func (cs *Table) SwapRemoveEntity(pos soa.BlockPos) (uid.UID64, bool) {
	lastChunkIdx, lastChunk := cs.ResolveTail()
	lastSlot := soa.ChunkSlot(lastChunk.Len - 1)
	targetChunk := &cs.Chunks[pos.ChunkIdx]

	if pos.ChunkIdx == lastChunkIdx && pos.ChunkSlot == lastSlot {
		cs.ZeroSlot(lastChunk, lastSlot)
		lastChunk.Len--
		cs.Len--
		return 0, false
	}

	entityToMove := *(*uid.UID64)(cs.GetEntityColumn().At(lastChunk, lastSlot))
	cs.SwapCopy(targetChunk, pos.ChunkSlot, lastChunk, lastSlot)
	cs.ZeroSlot(lastChunk, lastSlot)
	lastChunk.Len--
	cs.Len--

	return entityToMove, true
}

func (cs *Table) Clear() {
	cs.Block.Clear()
	cs.columns = nil
	cs.compColIdx.Reset()
}
