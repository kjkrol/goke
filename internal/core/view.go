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
	CompIDs     []ComponentID
	matched     []*Archetype
	Baked       []MatchedArch
}

// View factory based on Functional Options pattern
func NewView(blueprint *Blueprint, reg *Registry) *View {
	var mask ArchetypeMask
	var excludedMask ArchetypeMask

	for _, id := range blueprint.compIDs {
		mask = mask.Set(id)
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
		CompIDs:     blueprint.compIDs,
	}
	v.Reindex()
	v.Reg.ViewRegistry.Register(v)
	return v
}

func (v *View) Reindex() {
	v.matched = v.matched[:0]
	for _, arch := range v.Reg.ArchetypeRegistry.All() {
		if v.Matches(arch.Mask) {
			v.matched = append(v.matched, arch)
		}
	}

	v.Baked = v.Baked[:0]
	for _, arch := range v.matched {
		v.AddArchetype(arch)
	}
}

func (v *View) AddArchetype(arch *Archetype) {
	mArch := MatchedArch{
		Entities: &arch.entities,
		Len:      &arch.len,
	}
	for i, id := range v.CompIDs {
		mArch.Columns[i] = arch.Columns[id]
	}
	v.Baked = append(v.Baked, mArch)
}

func (v *View) Matches(archMask ArchetypeMask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}
