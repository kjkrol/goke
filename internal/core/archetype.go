package core

type Archetype struct {
	Mask     ArchetypeMask
	entities []Entity
	Columns  [MaxComponents]*Column
	// Cached IDs of active components for high-speed iteration
	activeIDs []ComponentID

	len     int
	cap     int
	initCap int

	edgesNext [MaxComponents]*Archetype
	edgesPrev [MaxComponents]*Archetype
}

type ArchRow uint32

func (a *Archetype) SwapRemoveEntity(row ArchRow) (swapedEntity Entity, swaped bool) {
	lastRow := ArchRow(a.len - 1)
	entityToMove := a.entities[lastRow]

	// 1. Swap data in all active columns using cached IDs
	for _, id := range a.activeIDs {
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

	if row == lastRow {
		return 0, false
	}
	return entityToMove, true
}

func (a *Archetype) registerEntity(entity Entity) ArchRow {
	a.ensureCapacity()
	newIdx := a.len

	// Update column lengths using cached IDs
	for _, id := range a.activeIDs {
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

	// Grow columns using cached IDs
	for _, id := range a.activeIDs {
		a.Columns[id].growTo(newCap)
	}

	a.cap = newCap
}

// CountNextEdges remains as is (or use a stored counter if needed)
func (a *Archetype) CountNextEdges() int {
	return countNonNull(a.edgesNext)
}

func (a *Archetype) CountPrevEdges() int {
	return countNonNull(a.edgesPrev)
}

func countNonNull(edges [MaxComponents]*Archetype) int {
	count := 0
	for _, edge := range edges {
		if edge != nil {
			count++
		}
	}
	return count
}
