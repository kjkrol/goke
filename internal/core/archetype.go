package core

// Supports 256 unique component types
type Archetype struct {
	Mask     ArchetypeMask
	entities []Entity
	Columns  [MaxComponents]*Column
	len      int
	cap      int

	edgesNext [MaxComponents]*Archetype
	edgesPrev [MaxComponents]*Archetype
	initCap   int
}

type ArchRow uint32

type EntityArchLink struct {
	Arch *Archetype
	Row  ArchRow
}

func NewArchetype(mask ArchetypeMask, defaultArchetypeChunkSize int) *Archetype {
	return &Archetype{
		Mask:     mask,
		entities: make([]Entity, defaultArchetypeChunkSize),
		len:      0,
		cap:      defaultArchetypeChunkSize,
		initCap:  defaultArchetypeChunkSize,
	}
}

func (a *Archetype) SwapRemoveEntity(row ArchRow) (swapedEntity Entity, swaped bool) {
	lastRow := ArchRow(a.len - 1)
	entityToMove := a.entities[lastRow]

	// 1. Swap data in all columns
	for id := range a.Mask.AllSet() {
		col := a.Columns[id]

		if row != lastRow {
			col.copyData(row, lastRow)
		}

		col.zeroData(lastRow)
		col.len--
	}

	// 2. Swap entity ID in the entities slice
	a.entities[row] = entityToMove
	a.entities[lastRow] = 0
	a.len--
	// 3. Return the entity that was moved to the new position
	// If we removed the last one, no entity was moved to 'index'
	if row == lastRow {
		return 0, false
	}
	return entityToMove, true
}

func (a *Archetype) registerEntity(entity Entity) ArchRow {
	a.ensureCapacity()
	newIdx := a.len

	for id := range a.Mask.AllSet() {
		a.Columns[id].len++
	}

	a.entities[newIdx] = entity
	a.len++

	return ArchRow(newIdx)
}

func (a *Archetype) ensureCapacity() {
	if a.len < a.cap {
		return
	}

	newCap := a.cap * 2
	if newCap == 0 {
		newCap = a.initCap
	}

	newEntities := make([]Entity, newCap)
	copy(newEntities, a.entities)
	a.entities = newEntities

	for id := range a.Mask.AllSet() {
		a.Columns[id].growTo(newCap)
	}

	a.cap = newCap
}

// CountNextEdges returns the number of outgoing connections (adding a component).
func (a *Archetype) CountNextEdges() int {
	return countNonNull(a.edgesNext)
}

// CountPrevEdges returns the number of incoming connections (removing a component).
func (a *Archetype) CountPrevEdges() int {
	return countNonNull(a.edgesPrev)
}

// Internal helper to iterate through the fixed-size array
func countNonNull(edges [MaskSize * 64]*Archetype) int {
	count := 0
	for _, edge := range edges {
		if edge != nil {
			count++
		}
	}
	return count
}
