//go:generate go run ./gen/main.go
package ecs

import "unsafe"

const MaxComponents = 8

type matchedArch struct {
	arch     *Archetype
	entities []Entity
	count    int
	ptrs     [MaxComponents]unsafe.Pointer
	sizes    [MaxComponents]uintptr
}

type View struct {
	reg     *Registry
	mask    ArchetypeMask
	ids     []ComponentID
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
	numIds := len(v.ids)
	for _, arch := range v.matched {
		if arch.len == 0 {
			continue
		}

		mArch := matchedArch{
			arch:     arch,
			entities: arch.entities[:arch.len],
			count:    arch.len,
		}

		for i := 0; i < numIds; i++ {
			col := arch.columns[v.ids[i]]
			mArch.ptrs[i] = col.data
			mArch.sizes[i] = col.itemSize
		}
		v.baked = append(v.baked, mArch)
	}
}
