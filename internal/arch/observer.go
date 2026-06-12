package arch

type ArchetypeObserver interface {
	OnArchetypeCreated(*Archetype)
}
