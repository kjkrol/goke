package ecs

type entitiesRegistry struct {
	gen     *entityGenerator
	masks   []ArchetypeMask
	records []entityRecord
}

func newEntitiesRegistry() *entitiesRegistry {
	const initialCapacity = 1024

	return &entitiesRegistry{
		gen:   newEntityGenerator(initialCapacity),
		masks: make([]ArchetypeMask, 0, initialCapacity),
	}
}

func (m *entitiesRegistry) GetRecord(e Entity) (*entityRecord, bool) {
	id := uint32(e & IndexMask)

	if int(id) >= len(m.records) {
		return &entityRecord{}, false
	}

	if m.gen.generations[id] != uint32(e>>GenerationShift) {
		return &entityRecord{}, false
	}

	return &m.records[id], true
}

func (m *entitiesRegistry) create() Entity {
	e := m.gen.next()
	index := uint32(uint64(e) & IndexMask)

	for int(index) >= len(m.masks) {
		m.records = append(m.records, entityRecord{})
		m.masks = append(m.masks, ArchetypeMask{})
	}

	m.masks[index] = ArchetypeMask{}
	return e
}

func (m *entitiesRegistry) destroy(e Entity) bool {
	if !m.isValid(e) {
		return false
	}

	index := uint32(uint64(e) & IndexMask)

	m.gen.release(e)
	m.masks[index] = ArchetypeMask{}

	return true
}

func (m *entitiesRegistry) GetMask(e Entity) (ArchetypeMask, bool) {
	index := uint32(uint64(e) & IndexMask)
	gen := uint32(uint64(e) >> GenerationShift)

	if index >= uint32(len(m.masks)) {
		return ArchetypeMask{}, false
	}

	if m.gen.generations[index] != gen {
		return ArchetypeMask{}, false
	}

	return m.masks[index], true
}

func (m *entitiesRegistry) updateMask(e Entity, newMask ArchetypeMask) bool {
	index := uint32(uint64(e) & IndexMask)
	gen := uint32(uint64(e) >> GenerationShift)

	if index >= uint32(len(m.masks)) || m.gen.generations[index] != gen {
		return false
	}

	m.masks[index] = newMask
	return true
}

func (m *entitiesRegistry) setRecord(e Entity, arch *archetype, index int) {
	id := uint32(e & IndexMask)
	m.records[id] = entityRecord{arch: arch, index: index}
}

func (m *entitiesRegistry) isValid(e Entity) bool {
	index := uint32(e & IndexMask)
	gen := uint32(e >> GenerationShift)
	return index < uint32(len(m.gen.generations)) && m.gen.generations[index] == gen
}
