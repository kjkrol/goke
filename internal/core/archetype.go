package core

import "unsafe"

// ArchetypeId represents a unique identifier within the archetype registry.
//
// Special values:
//   - 0: Indicates a non-existent archetype (Null Archetype).
//   - 1: Represents the Root Archetype, acting as the entry point of the graph.
//
// Archetypes are organized in a graph structure, where nodes are connected
// through edges (e.g., adding or removing components) to facilitate
// efficient entity transitions.
type ArchetypeId uint16

const NullArchetypeId = ArchetypeId(0)
const RootArchetypeId = ArchetypeId(1)
const MaxArchetypeId = ArchetypeId(4096)

type LocalColumnID uint8

const EntityColumnIndex = LocalColumnID(0)
const FirstDataColumnIndex = LocalColumnID(1)
const InvalidLocalID = LocalColumnID(MaxComponents + 1)

type Archetype struct {
	Mask ArchetypeMask
	Id   ArchetypeId

	// Global ComponentID -> Local Index in 'columns' slice
	// 128 bytes - fits in exactly 2 cache lines.
	columnMap [MaxComponents]LocalColumnID
	// Dense storage:
	// columns[0] = Entity IDs (Always present)
	// columns[1..N] = Component Data
	columns []Column

	len     int
	cap     int
	initCap int

	edgesNext [MaxComponents]ArchetypeId
	edgesPrev [MaxComponents]ArchetypeId
}

type ArchRow uint32

func (a *Archetype) SwapRemoveEntity(row ArchRow) (swapedEntity Entity, swaped bool) {
	lastRow := ArchRow(a.len - 1)
	entityCol := &a.columns[EntityColumnIndex]
	ptr := entityCol.Data
	stride := entityCol.ItemSize

	srcPtr := unsafe.Add(ptr, uintptr(lastRow)*stride)
	entityToMove := *(*Entity)(srcPtr)

	// 1. Swap data in all active columns using cached IDs
	for i := range a.columns {
		col := &a.columns[i] // Get pointer to struct in slice

		if row != lastRow {
			col.copyData(row, lastRow)
		}

		col.zeroData(lastRow)
		col.len--
	}

	// 2. Swap entity ID
	a.len--

	if row == lastRow {
		return 0, false
	}
	return entityToMove, true
}

func (a *Archetype) registerEntity(entity Entity) ArchRow {
	a.ensureCapacity()

	// 1. Write Entity ID to Column 0
	// We assume columns[0] is initialized with ItemSize = sizeof(Entity)
	entityCol := &a.columns[EntityColumnIndex]

	// Calculate address: Data + (len * 8)
	targetPtr := unsafe.Add(entityCol.Data, uintptr(a.len)*entityCol.ItemSize)

	// Store the entity ID
	*(*Entity)(targetPtr) = entity

	// 2. Increment length in ALL columns
	for i := range a.columns {
		a.columns[i].len++
	}

	newRow := ArchRow(a.len)
	a.len++

	return newRow
}

func (a *Archetype) ensureCapacity() {
	if a.len < a.cap {
		return
	}

	newCap := a.cap * 2
	if newCap == 0 {
		newCap = a.initCap
	}
	if newCap == 0 {
		newCap = 1
	}

	// Simplified: We just iterate over ALL columns.
	// Since Entity is now column[0], it gets grown automatically here.
	for i := range a.columns {
		a.columns[i].growTo(newCap)
	}

	a.cap = newCap
}

func (a *Archetype) GetEntityColumn() *Column {
	return &a.columns[EntityColumnIndex]
}

func (a *Archetype) GetColumn(id ComponentID) *Column {
	localIdx := a.columnMap[id]
	if localIdx == InvalidLocalID {
		return nil
	}
	return &a.columns[localIdx]
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
