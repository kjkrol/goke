package ecs

const MaxComponents = 8

type matchedArch struct {
	entities *[]Entity
	columns  [MaxComponents]*column
	len      *int
}

type View struct {
	reg         *Registry
	includeMask ArchetypeMask
	excludeMask ArchetypeMask
	compIDs     []ComponentID
	matched     []*Archetype
	baked       []matchedArch
}

func (v *View) Reindex() {
	v.matched = v.matched[:0]
	for _, arch := range v.reg.archetypeRegistry.All() {
		if v.matches(arch.mask) {
			v.matched = append(v.matched, arch)
		}
	}

	v.baked = v.baked[:0]
	for _, arch := range v.matched {
		v.AddArchetype(arch)
	}
}

func (v *View) AddArchetype(arch *Archetype) {
	mArch := matchedArch{
		entities: &arch.entities,
		len:      &arch.len,
	}
	for i, id := range v.compIDs {
		mArch.columns[i] = arch.columns[id]
	}
	v.baked = append(v.baked, mArch)
}

// ---------------------------------------------------------

func (v *View) matches(archMask ArchetypeMask) bool {
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
