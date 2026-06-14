package arch

type Observer interface {
	OnArchetypeCreated(*Archetype)
}
