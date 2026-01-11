package ecs

type Entity uint64

type entitiesRegistry struct {
	gen   *entityGenerator
	list  *aliveList
	masks []Bitmask
}

func newEntitiesRegistry() *entitiesRegistry {
	const initialCapacity = 1024

	return &entitiesRegistry{
		gen:   newEntityGenerator(initialCapacity),
		list:  newAliveList(initialCapacity),
		masks: make([]Bitmask, 0, initialCapacity),
	}
}

func (m *entitiesRegistry) create() Entity {
	e := m.gen.next()
	index := uint32(uint64(e) & IndexMask)

	for int(index) >= len(m.masks) {
		m.masks = append(m.masks, Bitmask{})
	}

	m.masks[index] = Bitmask{}
	m.list.add(index)
	return e
}

func (m *entitiesRegistry) destroy(e Entity) uint32 {
	index := m.gen.release(e)

	m.masks[index] = Bitmask{}
	m.list.remove(index)

	return index
}

func (m *entitiesRegistry) mask(e Entity) (Bitmask, bool) {
	index := uint32(uint64(e) & IndexMask)
	gen := uint32(uint64(e) >> GenerationShift)

	if index >= uint32(len(m.masks)) {
		return Bitmask{}, false
	}

	if m.gen.generations[index] != gen {
		return Bitmask{}, false
	}

	return m.masks[index], true
}

func (m *entitiesRegistry) updateMask(e Entity, newMask Bitmask) bool {
	index := uint32(uint64(e) & IndexMask)
	gen := uint32(uint64(e) >> GenerationShift)

	if index >= uint32(len(m.masks)) || m.gen.generations[index] != gen {
		return false
	}

	m.masks[index] = newMask
	return true
}

func (m *entitiesRegistry) active() func(func(e Entity, mask Bitmask) bool) {
	return func(yield func(e Entity, mask Bitmask) bool) {
		for _, index := range m.list.alive {
			gen := m.gen.generations[index]
			e := Entity(uint64(gen)<<GenerationShift | uint64(index))

			if !yield(e, m.masks[index]) {
				return
			}
		}
	}
}

//-----------------------------

type aliveList struct {
	alive   []uint32 // Sparse table
	indices []int    // map: ID -> pos in alive
}

func newAliveList(initialCapacity int) *aliveList {
	return &aliveList{
		alive:   make([]uint32, 0, initialCapacity),
		indices: make([]int, 0, initialCapacity),
	}
}

func (l *aliveList) add(index uint32) {
	for int(index) >= len(l.indices) {
		l.indices = append(l.indices, -1)
	}
	l.indices[index] = len(l.alive)
	l.alive = append(l.alive, index)
}

func (l *aliveList) remove(index uint32) {
	pos := l.indices[index]
	if pos == -1 {
		return
	}

	lastIdx := len(l.alive) - 1
	lastEntityIndex := l.alive[lastIdx]

	// Swap & Pop
	l.alive[pos] = lastEntityIndex
	l.indices[lastEntityIndex] = pos

	l.alive = l.alive[:lastIdx]
	l.indices[index] = -1
}

//-----------------------------

const (
	IndexMask       = 0xFFFFFFFF
	GenerationShift = 32
)

type entityGenerator struct {
	lastID      uint32
	freeList    []uint32
	generations []uint32
}

func newEntityGenerator(initialCapacity int) *entityGenerator {
	return &entityGenerator{
		generations: make([]uint32, 0, initialCapacity),
		freeList:    make([]uint32, 0, 128),
	}
}

func (g *entityGenerator) next() Entity {
	var index uint32
	if len(g.freeList) > 0 {
		index = g.freeList[len(g.freeList)-1]
		g.freeList = g.freeList[:len(g.freeList)-1]
	} else {
		index = g.lastID
		g.lastID++
		g.generations = append(g.generations, 0)
	}

	gen := g.generations[index]
	return Entity(uint64(gen)<<GenerationShift | uint64(index))
}

func (g *entityGenerator) release(e Entity) uint32 {
	index := uint32(uint64(e) & IndexMask)
	g.generations[index]++
	g.freeList = append(g.freeList, index)
	return index
}
