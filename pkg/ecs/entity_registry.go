package ecs

type entitiesRegistry struct {
	entityPool      *EntityGenerationalPool
	entityBacklinks []ArchRowBacklink
}

type ArchRowBacklink struct {
	arch  *archetype
	index int
}

func newEntitiesRegistry() *entitiesRegistry {
	const initialCapacity = 1024

	return &entitiesRegistry{
		entityPool:      NewEntityGenerator(initialCapacity),
		entityBacklinks: make([]ArchRowBacklink, 0, initialCapacity),
	}
}

func (m *entitiesRegistry) GetBackLink(e Entity) (*ArchRowBacklink, bool) {
	index := e.Index()

	if int(index) >= len(m.entityBacklinks) {
		return &ArchRowBacklink{}, false
	}

	if !m.entityPool.IsValid(e) {
		return &ArchRowBacklink{}, false
	}

	return &m.entityBacklinks[index], true
}

func (m *entitiesRegistry) create() Entity {
	entity := m.entityPool.Next()

	index := entity.Index()
	for int(index) >= len(m.entityBacklinks) {
		m.entityBacklinks = append(m.entityBacklinks, ArchRowBacklink{})
	}

	m.entityBacklinks[index] = ArchRowBacklink{}
	return entity
}

func (m *entitiesRegistry) destroy(e Entity) bool {
	if !m.entityPool.IsValid(e) {
		return false
	}

	m.entityPool.Release(e)
	index := e.Index()
	m.entityBacklinks[index] = ArchRowBacklink{}

	return true
}

func (m *entitiesRegistry) SetBacklink(e Entity, arch *archetype, indexInArch int) {
	index := e.Index()
	m.entityBacklinks[index] = ArchRowBacklink{arch: arch, index: indexInArch}
}
