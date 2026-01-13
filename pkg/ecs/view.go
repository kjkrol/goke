package ecs

type viewBase struct {
	reg     *Registry
	mask    ArchetypeMask
	matched []*archetype
}

func (v *viewBase) Reindex() {
	// Resetujemy listę pasujących archetypów, zachowując zaalokowaną pamięć
	v.matched = v.matched[:0]

	// Pobieramy wszystkie archetypy z rejestru
	allArchs := v.reg.archetypeRegistry.All()

	for _, arch := range allArchs {
		// Sprawdzamy, czy maska archetypu zawiera wszystkie wymagane komponenty widoku
		if arch.mask.Contains(v.mask) {
			v.matched = append(v.matched, arch)
		}
	}
}
