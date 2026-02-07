package core

// ArchetypeId represents a unique identifier within the archetype registry.
//
// Special values:
//   - 0: Indicates a non-existent archetype (Null Archetype).
//   - 1: Represents the Root Archetype, acting as the entry point of the graph.
//
// Archetypes are organized in a graph structure, where nodes are connected
// through edges (e.g., adding or removing components) to facilitate
// efficient entity transitions.
type ArchetypeId int

const NullArchetypeId = ArchetypeId(0)
const RootArchetypeId = ArchetypeId(1)

type Archetype struct {
	Mask     ArchetypeMask
	Id       ArchetypeId
	entities []Entity
	Columns  [MaxComponents]*Column
	// Cached IDs of active components for high-speed iteration
	activeIDs []ComponentID

	len     int
	cap     int
	initCap int

	edgesNext [MaxComponents]ArchetypeId
	edgesPrev [MaxComponents]ArchetypeId
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
	return countNonZeros(a.edgesNext)
}

func (a *Archetype) CountPrevEdges() int {
	return countNonZeros(a.edgesPrev)
}

func countNonZeros(edges [MaxComponents]ArchetypeId) int {
	count := 0
	for _, edge := range edges {
		if edge != 0 {
			count++
		}
	}
	return count
}
