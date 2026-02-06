package core

import (
	"fmt"
	"unsafe"
)

const MaxColumns = 8

type MatchedArch struct {
	base        uintptr
	offEntities uintptr
	offLen      uintptr

	colDataPtrs     [MaxColumns]uintptr
	colItemSizePtrs [MaxColumns]uintptr
}

func (v *View) AddArchetype(arch *Archetype) {
	mArch := MatchedArch{
		base:        uintptr(unsafe.Pointer(arch)),
		offEntities: unsafe.Offsetof(arch.entities),
		offLen:      unsafe.Offsetof(arch.len),
	}

	for i, info := range v.CompInfos {
		if i >= MaxColumns {
			break
		}

		col := arch.Columns[info.ID]
		if col != nil {
			mArch.colDataPtrs[i] = uintptr(unsafe.Pointer(&col.Data))
			mArch.colItemSizePtrs[i] = uintptr(unsafe.Pointer(&col.ItemSize))
		}
	}
	v.Baked = append(v.Baked, mArch)
}

func (m *MatchedArch) GetEntities() []Entity {
	return *(*[]Entity)(unsafe.Pointer(m.base + m.offEntities))
}

func (m *MatchedArch) GetData(idx int) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(m.colDataPtrs[idx]))
}

func (m *MatchedArch) GetItemSize(idx int) uintptr {
	return *(*uintptr)(unsafe.Pointer(m.colItemSizePtrs[idx]))
}

func (m *MatchedArch) GetLen() int {
	return *(*int)(unsafe.Pointer(m.base + m.offLen))
}

type View struct {
	Reg         *Registry
	includeMask ArchetypeMask
	excludeMask ArchetypeMask
	CompInfos   []ComponentInfo
	Baked       []MatchedArch
}

// View factory based on Functional Options pattern
func NewView(blueprint *Blueprint, reg *Registry) *View {
	var mask ArchetypeMask
	var excludedMask ArchetypeMask

	for _, info := range blueprint.compInfos {
		mask = mask.Set(info.ID)
	}

	for _, id := range blueprint.tagIDs {
		mask = mask.Set(id)
	}

	for _, id := range blueprint.exCompIDs {
		if mask.IsSet(id) {
			panic(fmt.Sprintf("ECS View Error: Component ID %d cannot be both REQUIRED and EXCLUDED", id))
		}
		excludedMask = excludedMask.Set(id)
	}

	v := &View{
		Reg:         reg,
		includeMask: mask,
		excludeMask: excludedMask,
		CompInfos:   blueprint.compInfos,
	}
	v.Reindex()
	v.Reg.ViewRegistry.Register(v)
	return v
}

func (v *View) Reindex() {
	v.Baked = v.Baked[:0]
	for _, arch := range v.Reg.ArchetypeRegistry.All() {
		if v.Matches(arch.Mask) {
			v.AddArchetype(arch)
		}
	}
}

func (v *View) Matches(archMask ArchetypeMask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}
