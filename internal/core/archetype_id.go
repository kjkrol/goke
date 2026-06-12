package core

type ArchetypeId uint16

const (
	NullArchetypeId = ArchetypeId(0)
	RootArchetypeId = ArchetypeId(1)
	MaxArchetypeId  = ArchetypeId(4096)
)
