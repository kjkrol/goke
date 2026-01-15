package ecs

type EntityGenerationalPool struct {
	lastIndex    uint32
	freeIndicies []uint32
	generations  []uint32
}

func NewEntityGenerator(initialCapacity int) *EntityGenerationalPool {
	return &EntityGenerationalPool{
		generations:  make([]uint32, 0, initialCapacity),
		freeIndicies: make([]uint32, 0, 128),
	}
}

func (p *EntityGenerationalPool) Next() Entity {
	var index uint32
	if len(p.freeIndicies) > 0 {
		index = p.freeIndicies[len(p.freeIndicies)-1]
		p.freeIndicies = p.freeIndicies[:len(p.freeIndicies)-1]
	} else {
		index = p.lastIndex
		p.lastIndex++
		p.generations = append(p.generations, 0)
	}

	gen := p.generations[index]
	return EntityFrom(gen, index)
}

func (p *EntityGenerationalPool) Release(e Entity) uint32 {
	index := IndexOf(e)
	p.generations[index]++
	p.freeIndicies = append(p.freeIndicies, index)
	return index
}

func (p *EntityGenerationalPool) IsValid(e Entity) bool {
	index, gen := IndexWithGenOf(e)
	return index < uint32(len(p.generations)) && p.generations[index] == gen
}
