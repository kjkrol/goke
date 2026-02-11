package core

import (
	"fmt"
)

type MatchedArch struct {
	Arch            *Archetype
	LayoutColumnIDs []LocalColumnID
}

func (ma *MatchedArch) GetEntityColumn() *Column {
	return &ma.Arch.columns[0]
}

func (ma *MatchedArch) GetColumn(colId LocalColumnID) *Column {
	layoutColId := ma.LayoutColumnIDs[colId]
	return &ma.Arch.columns[layoutColId]
}

type View struct {
	Reg         *Registry
	includeMask ArchetypeMask
	excludeMask ArchetypeMask
	Layout      []ComponentInfo
	Baked       []MatchedArch
}

// View factory based on Functional Options pattern
func NewView(blueprint *Blueprint, layout []ComponentInfo, reg *Registry) *View {
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

	for _, info := range layout {
		if !mask.IsSet(info.ID) {
			panic(fmt.Sprintf("View Layout Error: Component %d is in layout but not required by Blueprint", info.ID))
		}
	}

	v := &View{
		Reg:         reg,
		includeMask: mask,
		excludeMask: excludedMask,
		Layout:      layout,
	}
	v.Reindex()
	v.Reg.ViewRegistry.Register(v)
	return v
}

func (v *View) Reindex() {
	v.Baked = v.Baked[:0]
	reg := v.Reg.ArchetypeRegistry

	for i := RootArchetypeId; i < reg.lastArchetypeId; i++ {
		arch := &reg.Archetypes[i]

		if len(arch.columns) == 0 {
			continue
		}

		if v.Matches(arch.Mask) {
			v.AddArchetype(arch)
		}
	}
}

func (v *View) AddArchetype(arch *Archetype) {

	if len(arch.columns) == 0 {
		return
	}

	mArch := MatchedArch{
		Arch: arch,
	}

	if len(v.Layout) > 0 {
		mArch.LayoutColumnIDs = make([]LocalColumnID, len(v.Layout))

		for i, info := range v.Layout {
			localIdx := arch.columnMap[info.ID]
			if localIdx == InvalidLocalID || int(localIdx) >= len(arch.columns) {
				continue
			}
			mArch.LayoutColumnIDs[i] = localIdx
		}
	}

	v.Baked = append(v.Baked, mArch)
}

func (v *View) Matches(archMask ArchetypeMask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}
