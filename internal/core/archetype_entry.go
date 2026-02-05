package core

import "unsafe"

const ArchetypeEntryCap = 8

type ArchetypeEntryBlueprint struct {
	reg     *Registry
	mask    ArchetypeMask
	CompIDs []ComponentID
	Arch    *Archetype
}

func NewArchetypeEntry(blueprint *Blueprint) *ArchetypeEntryBlueprint {
	var mask ArchetypeMask

	for _, id := range blueprint.compIDs {
		mask = mask.Set(id)
	}

	for _, id := range blueprint.tagIDs {
		mask = mask.Set(id)
	}

	arch := blueprint.Reg.ArchetypeRegistry.getOrRegister(mask)

	return &ArchetypeEntryBlueprint{
		reg:     blueprint.Reg,
		mask:    mask,
		CompIDs: blueprint.compIDs,
		Arch:    arch,
	}
}

func (b *ArchetypeEntryBlueprint) Create() (Entity, [ArchetypeEntryCap]unsafe.Pointer) {
	entity := b.reg.EntityPool.Next()
	row := b.reg.ArchetypeRegistry.addEntity(entity, b.Arch)
	var ptrs [ArchetypeEntryCap]unsafe.Pointer
	for i, compId := range b.CompIDs {
		column := b.Arch.Columns[compId]
		ptrs[i] = unsafe.Add(column.Data, uintptr(row)*column.ItemSize)
	}
	return entity, ptrs
}
