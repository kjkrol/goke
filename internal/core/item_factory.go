package core

import (
	"sync"
	"unsafe"
)

const (
	archetypeEntryCap = 1 + 10 // 1 entity 10 components
	InitBatchSize     = 1024
	MaxBatchSize      = 8192
)

type Item [archetypeEntryCap]unsafe.Pointer

var bufPool = sync.Pool{
	New: func() any {
		return make([]Item, InitBatchSize)
	},
}

type ItemFactory struct {
	reg       *Registry
	mask      ArchetypeMask
	CompInfos []ComponentInfo
	ArchId    ArchetypeId
}

func NewItemFactory(blueprint *Blueprint) *ItemFactory {
	var mask ArchetypeMask

	for _, info := range blueprint.compInfos {
		mask = mask.Set(info.ID)
	}

	for _, tag := range blueprint.tagIDs {
		mask = mask.Set(tag)
	}

	archId := blueprint.Reg.ArchetypeRegistry.getOrRegister(mask)

	return &ItemFactory{
		reg:       blueprint.Reg,
		mask:      mask,
		CompInfos: blueprint.compInfos,
		ArchId:    archId,
	}
}

func (b *ItemFactory) Create() Item {
	// entity := b.reg.EntityPool.Next()
	// chunkIdx, chunkRow := b.reg.ArchetypeRegistry.AddEntity(entity, b.ArchId)
	// arch := b.reg.ArchetypeRegistry.Archetypes[b.ArchId]
	// chunk := arch.Memory.GetChunk(chunkIdx)
	// var ptrs Item
	// for i, info := range b.CompInfos {
	// 	column := arch.GetColumn(info.ID)
	// 	ptrs[i] = column.GetPointer(chunk, chunkRow)
	// }
	// return entity, ptrs
	var localBuf [1]Item
	b.batchCreate(1, localBuf[:])
	return localBuf[0]
}

func (b *ItemFactory) BatchCreate(count int, fn func([]Item)) {
	buf := bufPool.Get().([]Item)

	defer func() {
		if cap(buf) <= MaxBatchSize {
			bufPool.Put(buf)
		}
	}()

	if cap(buf) < count {
		buf = make([]Item, count)
	}

	buf = b.batchCreate(count, buf)

	fn(buf)
}

func (b *ItemFactory) batchCreate(count int, buf []Item) []Item {
	output := buf[:count]

	arch := &b.reg.ArchetypeRegistry.Archetypes[b.ArchId]
	memo := &arch.Memory
	entityCol := &arch.Columns[EntityColumnIndex]

	var columns [archetypeEntryCap]*Column
	for i, info := range b.CompInfos {
		columns[i] = arch.GetColumn(info.ID)
	}

	remaining := count
	outIdx := 0

	for remaining > 0 {
		chunk, chunkIdx, startRow, allocatedRows := memo.AllocBatch(remaining)

		for i := range allocatedRows {
			entity := b.reg.EntityPool.Next()
			rowIdx := startRow + PageRow(i)

			destPtr := entityCol.GetPointer(chunk, rowIdx)
			*(*Entity)(destPtr) = entity
			b.reg.ArchetypeRegistry.EntityLinkStore.Update(entity, b.ArchId, chunkIdx, rowIdx)

			output[outIdx][0] = destPtr

			for j := range b.CompInfos {
				output[outIdx][j+1] = columns[j].GetPointer(chunk, rowIdx)
			}

			outIdx++
		}

		remaining -= allocatedRows
	}

	return output
}
