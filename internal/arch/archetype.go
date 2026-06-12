package arch

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/internal/mem"
)

type Archetype struct {
	Mask    core.ArchetypeMask
	Id      core.ArchetypeId
	Map     mem.ColumnMap
	Memory  mem.Memo
	Columns []mem.Column
	graph   *ArchetypeGraph
}

func (a *Archetype) Reset() {
	a.Memory.Clear()
	if a.graph != nil {
		a.graph.Reset()
	}
	a.Map.Reset()
	a.Mask = core.ArchetypeMask{}
	a.Id = core.NullArchetypeId
	a.Columns = nil
}

func (a *Archetype) InitArchetype(
	archId core.ArchetypeId,
	mask core.ArchetypeMask,
	colsInfos []core.ComponentInfo,
) {
	a.Id = archId
	a.Mask = mask
	a.graph = &ArchetypeGraph{}

	a.Memory.Init(colsInfos)

	count := len(colsInfos) + 1
	a.Columns = make([]mem.Column, count)
	a.Map.Reset()

	a.Columns[mem.EntityColumnIndex] = mem.Column{
		CompID:     core.EntityID,
		ItemSize:   unsafe.Sizeof(uid.UID64(0)),
		PageOffset: a.Memory.Layout.Offsets[0],
	}

	for i, info := range colsInfos {
		localIdx := mem.LocalColumnID(i + 1)
		a.Map.Set(info.ID, localIdx)
		a.Columns[localIdx] = mem.Column{
			CompID:     info.ID,
			ItemSize:   info.Size,
			PageOffset: a.Memory.Layout.Offsets[i+1],
		}
	}
}

func (a *Archetype) Len() int {
	return int(a.Memory.Len)
}

func (a *Archetype) GetEntityColumn() *mem.Column {
	return &a.Columns[mem.EntityColumnIndex]
}

func (a *Archetype) GetColumn(id core.ComponentID) *mem.Column {
	localIdx := a.Map.Get(id)
	if localIdx == mem.InvalidLocalID {
		return nil
	}
	return &a.Columns[localIdx]
}

func (a *Archetype) AddEntity(entity uid.UID64) (mem.PageIdx, mem.PageSlot) {
	page, pageIdx, pageSlot := a.Memory.AllocSlot()
	entityCol := &a.Columns[mem.EntityColumnIndex]
	destPtr := entityCol.GetPointer(page, pageSlot)
	*(*uid.UID64)(destPtr) = entity
	return pageIdx, pageSlot
}

func (a *Archetype) SwapRemoveEntity(targetChunkIdx mem.PageIdx, targetRow mem.PageSlot) (swappedEntity uid.UID64, swapped bool) {
	lastChunkIdx, lastChunk := a.Memory.ResolveTail()
	lastRow := mem.PageSlot(lastChunk.Len - 1)
	targetChunk := &a.Memory.Pages[targetChunkIdx]

	if targetChunkIdx == lastChunkIdx && targetRow == lastRow {
		a.zeroEntityAt(lastChunk, lastRow)
		lastChunk.Len--
		a.Memory.Len--
		return 0, false
	}

	ptrLastEntity := a.GetEntityColumn().GetPointer(lastChunk, lastRow)
	entityToMove := *(*uid.UID64)(ptrLastEntity)

	for i := range a.Columns {
		col := &a.Columns[i]
		src := col.GetPointer(lastChunk, lastRow)
		dst := col.GetPointer(targetChunk, targetRow)
		copy(unsafe.Slice((*byte)(dst), col.ItemSize), unsafe.Slice((*byte)(src), col.ItemSize))
	}

	a.zeroEntityAt(lastChunk, lastRow)
	lastChunk.Len--
	a.Memory.Len--

	return entityToMove, true
}

func (a *Archetype) zeroEntityAt(page *mem.Page, pageSlot mem.PageSlot) {
	for i := range a.Columns {
		col := &a.Columns[i]
		ptr := col.GetPointer(page, pageSlot)
		mem.ZeroMemory(ptr, col.ItemSize)
	}
}

func (a *Archetype) linkNextArch(nextArch *Archetype, compID core.ComponentID) {
	a.graph.edgesNext[compID] = nextArch.Id
	nextArch.graph.edgesPrev[compID] = a.Id
}

func (a *Archetype) linkPrevArch(prevArch *Archetype, compID core.ComponentID) {
	a.graph.edgesPrev[compID] = prevArch.Id
	prevArch.graph.edgesNext[compID] = a.Id
}
