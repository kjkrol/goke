package ecs

type EntityGenerationalPool struct {
	lastIndex   uint32
	freeIndices []uint32
	generations []uint32
}

func NewEntityGenerator(initialCapacity int) *EntityGenerationalPool {
	return &EntityGenerationalPool{
		generations: make([]uint32, 0, initialCapacity),
		freeIndices: make([]uint32, 0, 128),
	}
}

func (p *EntityGenerationalPool) Next() Entity {
	var index uint32
	if len(p.freeIndices) > 0 {
		index = p.freeIndices[len(p.freeIndices)-1]
		p.freeIndices = p.freeIndices[:len(p.freeIndices)-1]
	} else {
		index = p.lastIndex
		p.lastIndex++
		p.generations = append(p.generations, 0)
	}

	gen := p.generations[index]
	return NewEntity(gen, index)
}

func (p *EntityGenerationalPool) Release(e Entity) uint32 {
	index := e.Index()
	p.generations[index]++
	p.freeIndices = append(p.freeIndices, index)
	return index
}

func (p *EntityGenerationalPool) IsValid(e Entity) bool {
	if e.IsVirtual() {
		return false
	}
	index, gen := e.Unpack()
	return index < uint32(len(p.generations)) && p.generations[index] == gen
}
