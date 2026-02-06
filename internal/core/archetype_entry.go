package core

import "unsafe"

const ArchetypeEntryCap = 8

type ArchetypeEntryBlueprint struct {
	reg       *Registry
	mask      ArchetypeMask
	CompInfos []ComponentInfo
	Arch      *Archetype
}

func NewArchetypeEntry(blueprint *Blueprint) *ArchetypeEntryBlueprint {
	var mask ArchetypeMask

	for _, info := range blueprint.compInfos {
		mask = mask.Set(info.ID)
	}

	for _, tag := range blueprint.tagIDs {
		mask = mask.Set(tag)
	}

	var arch *Archetype = &Archetype{}
	blueprint.Reg.ArchetypeRegistry.getOrRegister(mask, &arch)

	return &ArchetypeEntryBlueprint{
		reg:       blueprint.Reg,
		mask:      mask,
		CompInfos: blueprint.compInfos,
		Arch:      arch,
	}
}

func (b *ArchetypeEntryBlueprint) Create() (Entity, [ArchetypeEntryCap]unsafe.Pointer) {
	entity := b.reg.EntityPool.Next()
	row := b.reg.ArchetypeRegistry.addEntity(entity, b.Arch)
	var ptrs [ArchetypeEntryCap]unsafe.Pointer
	for i, info := range b.CompInfos {
		column := b.Arch.Columns[info.ID]
		ptrs[i] = unsafe.Add(column.Data, uintptr(row)*column.ItemSize)
	}
	return entity, ptrs
}
