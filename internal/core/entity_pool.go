package core

type EntityGenerationalPool struct {
	lastIndex   uint32
	freeIndices []uint32
	generations []uint32
	capacity    uint32
}

func NewEntityGenerator(initialEntityCap, freeIndicesCap int) EntityGenerationalPool {
	return EntityGenerationalPool{
		// Pre-allocate with full length to avoid 'append' logic for generations
		generations: make([]uint32, initialEntityCap),
		// Pre-allocate capacity for freeIndices, but keep length 0
		freeIndices: make([]uint32, 0, freeIndicesCap),
		capacity:    uint32(initialEntityCap),
	}
}

func (p *EntityGenerationalPool) Next() Entity {
	// Priority 1: Reuse deleted indices (Fast Path)
	if fLen := len(p.freeIndices); fLen > 0 {
		index := p.freeIndices[fLen-1]
		p.freeIndices = p.freeIndices[:fLen-1]
		gen := p.generations[index]
		return NewEntity(gen, index)
	}

	// Priority 2: Use new index
	if p.lastIndex >= p.capacity {
		p.grow()
	}

	index := p.lastIndex
	p.lastIndex++

	// generations[index] is already 0 due to make() or grow() zero-init
	return NewEntity(p.generations[index], index)
}

func (p *EntityGenerationalPool) grow() {
	newCap := p.capacity * 2
	if newCap == 0 {
		newCap = 8 // Default fallback
	}

	// Grow generations - this is a heavy operation, but happens rarely
	newGenerations := make([]uint32, newCap)
	copy(newGenerations, p.generations)
	p.generations = newGenerations
	p.capacity = newCap
}

func (p *EntityGenerationalPool) Release(e Entity) uint32 {
	index := e.Index()
	// Increment generation to invalidate existing handles
	p.generations[index]++

	// We assume freeIndicesCap was set correctly to avoid grow here
	p.freeIndices = append(p.freeIndices, index)
	return index
}

func (p *EntityGenerationalPool) IsValid(e Entity) bool {
	index, gen := e.Unpack()
	return index < p.lastIndex && p.generations[index] == gen
}
