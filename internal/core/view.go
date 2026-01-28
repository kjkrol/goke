package core

const MaxComponents = 8

type MatchedArch struct {
	Entities *[]Entity
	Columns  [MaxComponents]*Column
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

// ---------------------------------------------------------

func (v *View) Matches(archMask ArchetypeMask) bool {
	// Inclusion - Unrolled with Early Exit
	if (archMask[0] & v.includeMask[0]) != v.includeMask[0] {
		return false
	}
	if (archMask[1] & v.includeMask[1]) != v.includeMask[1] {
		return false
	}
	if (archMask[2] & v.includeMask[2]) != v.includeMask[2] {
		return false
	}
	if (archMask[3] & v.includeMask[3]) != v.includeMask[3] {
		return false
	}

	// Exclusion - Unrolled with Early Exit
	if (archMask[0] & v.excludeMask[0]) != 0 {
		return false
	}
	if (archMask[1] & v.excludeMask[1]) != 0 {
		return false
	}
	if (archMask[2] & v.excludeMask[2]) != 0 {
		return false
	}
	if (archMask[3] & v.excludeMask[3]) != 0 {
		return false
	}

	return true
}
