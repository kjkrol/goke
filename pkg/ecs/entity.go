package ecs

const (
	IndexMask       = 0xFFFFFFFF
	GenerationShift = 32
)

type Entity uint64

//-----------------------------

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
