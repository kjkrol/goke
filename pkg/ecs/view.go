//go:generate go run ./gen/main.go
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
	compIDsLen := len(v.compIDs)
	for _, arch := range v.matched {
		if arch.len == 0 {
			continue
		}

		mArch := matchedArch{
			entities: &arch.entities,
			len:      &arch.len,
		}

		for i := range compIDsLen {
			col := arch.columns[v.compIDs[i]]
			mArch.columns[i] = col
		}
		v.baked = append(v.baked, mArch)
	}
}
