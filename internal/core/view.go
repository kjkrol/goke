package core

import "fmt"

const MaxColumns = 8

type MatchedArch struct {
	Entities *[]Entity
	Columns  [MaxColumns]*Column
	Len      *int
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
	arches := v.Reg.ArchetypeRegistry.Archetypes
	for i := range len(arches) {
		arch := &arches[i]
		if v.Matches(arch.Mask) {
			v.AddArchetype(arch)
		}
	}
}

func (v *View) AddArchetype(arch *Archetype) {
	mArch := MatchedArch{
		Entities: &arch.entities,
		Len:      &arch.len,
	}
	for i, info := range v.CompInfos {
		mArch.Columns[i] = arch.Columns[info.ID]
	}
	v.Baked = append(v.Baked, mArch)
}

func (v *View) Matches(archMask ArchetypeMask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}
