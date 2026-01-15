package ecs

type entitiesRegistry struct {
	entityPool *EntityGenerationalPool
	masks      []ArchetypeMask
	records    []ArchRowBacklink
}

type ArchRowBacklink struct {
	arch  *archetype
	index int
}

func newEntitiesRegistry() *entitiesRegistry {
	const initialCapacity = 1024

	return &entitiesRegistry{
		entityPool: NewEntityGenerator(initialCapacity),
		masks:      make([]ArchetypeMask, 0, initialCapacity),
	}
}

func (m *entitiesRegistry) GetRecord(e Entity) (*ArchRowBacklink, bool) {
	index, gen := IndexWithGenOf(e)

	if int(index) >= len(m.records) {
		return &ArchRowBacklink{}, false
	}

	if m.entityPool.generations[index] != gen {
		return &ArchRowBacklink{}, false
	}

	return &m.records[index], true
}

func (m *entitiesRegistry) create() Entity {
	e := m.entityPool.Next()

	index := IndexOf(e)
	for int(index) >= len(m.masks) {
		m.records = append(m.records, ArchRowBacklink{})
		m.masks = append(m.masks, ArchetypeMask{})
	}

	m.masks[index] = ArchetypeMask{}
	return e
}

func (m *entitiesRegistry) destroy(e Entity) bool {
	if !m.entityPool.IsValid(e) {
		return false
	}

	m.entityPool.Release(e)
	index := IndexOf(e)
	m.masks[index] = ArchetypeMask{}

	return true
}

func (m *entitiesRegistry) GetMask(e Entity) (ArchetypeMask, bool) {
	index, gen := IndexWithGenOf(e)

	if index >= uint32(len(m.masks)) {
		return ArchetypeMask{}, false
	}

	if m.entityPool.generations[index] != gen {
		return ArchetypeMask{}, false
	}

	return m.masks[index], true
}

func (m *entitiesRegistry) updateMask(e Entity, newMask ArchetypeMask) bool {
	index, gen := IndexWithGenOf(e)

	if index >= uint32(len(m.masks)) || m.entityPool.generations[index] != gen {
		return false
	}

	m.masks[index] = newMask
	return true
}

func (m *entitiesRegistry) SetBacklink(e Entity, arch *archetype, indexInArch int) {
	index := IndexOf(e)
	m.records[index] = ArchRowBacklink{arch: arch, index: indexInArch}
}
