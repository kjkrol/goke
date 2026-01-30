package core

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
