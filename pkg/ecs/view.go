package ecs

type viewBase struct {
	reg     *Registry
	mask    ArchetypeMask
	matched []*archetype
}

func (v *viewBase) Reindex() {
	v.matched = v.matched[:0]
	for _, arch := range v.reg.archetypeRegistry.All() {
		if arch.mask.Contains(v.mask) {
			v.matched = append(v.matched, arch)
		}
	}
}
