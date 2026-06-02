package core

import (
	"unsafe"
)

const (
	ArchetypeEntryCap = 10
	InitBatchSize     = 1024
)

type ArchetypeEntryBlueprint struct {
	reg       *Registry
	mask      ArchetypeMask
	CompInfos []ComponentInfo
	ArchId    ArchetypeId
	Buf       []ArchetypeBatchItem
}

func NewArchetypeEntry(blueprint *Blueprint) *ArchetypeEntryBlueprint {
	var mask ArchetypeMask

	for _, info := range blueprint.compInfos {
		mask = mask.Set(info.ID)
	}

	for _, tag := range blueprint.tagIDs {
		mask = mask.Set(tag)
	}

	archId := blueprint.Reg.ArchetypeRegistry.getOrRegister(mask)

	return &ArchetypeEntryBlueprint{
		reg:       blueprint.Reg,
		mask:      mask,
		CompInfos: blueprint.compInfos,
		ArchId:    archId,
		Buf:       make([]ArchetypeBatchItem, InitBatchSize),
	}
}

func (b *ArchetypeEntryBlueprint) Create() (Entity, [ArchetypeEntryCap]unsafe.Pointer) {

	entity := b.reg.EntityPool.Next()
	chunkIdx, chunkRow := b.reg.ArchetypeRegistry.AddEntity(entity, b.ArchId)
	arch := b.reg.ArchetypeRegistry.Archetypes[b.ArchId]
	chunk := arch.Memory.GetChunk(chunkIdx)
	var ptrs [ArchetypeEntryCap]unsafe.Pointer
	for i, info := range b.CompInfos {
		column := arch.GetColumn(info.ID)
		ptrs[i] = column.GetPointer(chunk, chunkRow)
	}
	return entity, ptrs
}

type ArchetypeBatchItem struct {
	Entity Entity
	Ptrs   [ArchetypeEntryCap]unsafe.Pointer
}

func (b *ArchetypeEntryBlueprint) CreateBatch(count int) []ArchetypeBatchItem {
	if cap(b.Buf) < count {
		b.Buf = make([]ArchetypeBatchItem, count)
	}
	output := b.Buf[:count]

	arch := &b.reg.ArchetypeRegistry.Archetypes[b.ArchId]
	memo := arch.Memory
	entityCol := &arch.Columns[EntityColumnIndex]

	var columns [ArchetypeEntryCap]*Column
	for i, info := range b.CompInfos {
		columns[i] = arch.GetColumn(info.ID)
	}

	remaining := count
	outIdx := 0

	for remaining > 0 {
		chunk, chunkIdx, startRow, allocatedRows := memo.AllocBatch(remaining)

		for i := range allocatedRows {
			entity := b.reg.EntityPool.Next()
			rowIdx := startRow + ChunkRow(i)

			destPtr := entityCol.GetPointer(chunk, rowIdx)
			*(*Entity)(destPtr) = entity
			b.reg.ArchetypeRegistry.EntityLinkStore.Update(entity, b.ArchId, chunkIdx, rowIdx)

			item := ArchetypeBatchItem{
				Entity: entity,
			}

			for j := range b.CompInfos {
				item.Ptrs[j] = columns[j].GetPointer(chunk, rowIdx)
			}

			// output = append(output, item)

			output[outIdx].Entity = entity

			for j := range b.CompInfos {
				output[outIdx].Ptrs[j] = columns[j].GetPointer(chunk, rowIdx)
			}

			outIdx++
		}

		remaining -= allocatedRows
	}

	// b.Buf = output
	return output
}
