package core

import "unsafe"

const ArchetypeEntryCap = 8

type ArchetypeEntryBlueprint struct {
	reg       *Registry
	mask      ArchetypeMask
	CompInfos []ComponentInfo
	ArchId    ArchetypeId
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
	}
}

func (b *ArchetypeEntryBlueprint) Create() (Entity, [ArchetypeEntryCap]unsafe.Pointer) {
	entity := b.reg.EntityPool.Next()
	row := b.reg.ArchetypeRegistry.AddEntity(entity, b.ArchId)
	arch := b.reg.ArchetypeRegistry.Archetypes[b.ArchId]
	var ptrs [ArchetypeEntryCap]unsafe.Pointer
	for i, info := range b.CompInfos {
		column := arch.GetColumn(info.ID)
		ptrs[i] = unsafe.Add(column.Data, uintptr(row)*column.ItemSize)
	}
	return entity, ptrs
}
