package core

import "unsafe"

const BlueprintCap = 8

type Blueprint struct {
	reg     *Registry
	mask    ArchetypeMask
	CompIDs []ComponentID
	Arch    *Archetype
}

func (b *Blueprint) Create() (Entity, [BlueprintCap]unsafe.Pointer) {
	entity := b.reg.EntityPool.Next()
	row := b.reg.ArchetypeRegistry.addEntity(entity, b.Arch)
	var ptrs [BlueprintCap]unsafe.Pointer
	for i, compId := range b.CompIDs {
		column := b.Arch.Columns[compId]
		ptrs[i] = unsafe.Add(column.Data, uintptr(row)*column.ItemSize)
	}
	return entity, ptrs
}

type BlueprintBuilder struct {
	reg     *Registry
	compIDs []ComponentID
	tagIDs  []ComponentID
}

type BlueprintOption func(*BlueprintBuilder)

func WithBlueprintTag[T any]() BlueprintOption {
	return func(b *BlueprintBuilder) {
		compInfo := EnsureComponentRegistered[T](b.reg.ComponentsRegistry)
		b.WithTag(compInfo.ID)
	}
}

func NewBlueprintBuilder(reg *Registry) *BlueprintBuilder {
	return &BlueprintBuilder{
		reg:     reg,
		compIDs: make([]ComponentID, 0, 8),
		tagIDs:  make([]ComponentID, 0, 8),
	}
}

func (b *BlueprintBuilder) WithTag(tagId ComponentID) {
	b.tagIDs = append(b.tagIDs, tagId)
}

func (b *BlueprintBuilder) WithComp(compId ComponentID) {
	b.compIDs = append(b.compIDs, compId)
}

func (b *BlueprintBuilder) Build() *Blueprint {
	var mask ArchetypeMask

	for _, id := range b.compIDs {
		mask = mask.Set(id)
	}

	for _, id := range b.tagIDs {
		mask = mask.Set(id)
	}

	arch := b.reg.ArchetypeRegistry.getOrRegister(mask)

	return &Blueprint{
		reg:     b.reg,
		mask:    mask,
		CompIDs: b.compIDs,
		Arch:    arch,
	}
}
