package ecs

const MaxComponents = 8

type matchedArch struct {
	entities *[]Entity
	columns  [MaxComponents]*column
	len      *int
}

type View struct {
	reg     *Registry
	mask    ArchetypeMask
	compIDs []ComponentID
	matched []*Archetype
	baked   []matchedArch
}

func (v *View) Reindex() {
	v.matched = v.matched[:0]
	for _, arch := range v.reg.archetypeRegistry.All() {
		if arch.mask.Contains(v.mask) {
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
