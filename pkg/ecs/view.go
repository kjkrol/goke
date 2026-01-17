//go:generate go run ./gen/main.go
package ecs

type viewBase struct {
	reg     *Registry
	mask    ArchetypeMask
	matched []*Archetype
}

func (v *viewBase) Reindex() {
	v.matched = v.matched[:0]
	for _, arch := range v.reg.archetypeRegistry.All() {
		if arch.mask.Contains(v.mask) {
			v.matched = append(v.matched, arch)
		}
	}
}

func (v *viewBase) AddTag(id ComponentID) {
	v.mask = v.mask.Set(id)
	v.Reindex()
}

func WithTag[T any](v *viewBase) {
	id := ensureComponentRegistered[T](v.reg.componentsRegistry)
	v.AddTag(id)
}
