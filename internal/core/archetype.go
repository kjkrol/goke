package core

import (
	"unsafe"
)

// -----------------------------------------------------------------------------
// ID & Constants
// -----------------------------------------------------------------------------

type ArchetypeId uint16

const (
	NullArchetypeId = ArchetypeId(0)
	RootArchetypeId = ArchetypeId(1)
	MaxArchetypeId  = ArchetypeId(4096)
)

type LocalColumnID uint8

const (
	EntityColumnIndex    = LocalColumnID(0)
	FirstDataColumnIndex = LocalColumnID(1)
	InvalidLocalID       = LocalColumnID(MaxComponents + 1)
)

type ArchRow uint32

// -----------------------------------------------------------------------------
// Column Map
// -----------------------------------------------------------------------------

// Global ComponentID -> Local Index in 'columns' slice inside MemoryBlock.
// 128 bytes - fits in exactly 2 cache lines.
type ColumnMap [MaxComponents]LocalColumnID

// Reset fills the map with InvalidLocalID.
func (m *ColumnMap) Reset() {
	for i := range m {
		m[i] = InvalidLocalID
	}
}

func (m *ColumnMap) Set(globalID ComponentID, localIdx LocalColumnID) {
	m[globalID] = localIdx
}

func (m *ColumnMap) Get(globalID ComponentID) LocalColumnID {
	return m[globalID]
}

// -----------------------------------------------------------------------------
// Archetype
// -----------------------------------------------------------------------------

type Archetype struct {
	Mask ArchetypeMask
	Id   ArchetypeId
	Map  ColumnMap

	// MemoryBlock is embedded by value for better cache locality & GC performance.
	// It manages the physical memory for all entities in this archetype.
	block MemoryBlock

	initCap int
	graph   *ArchetypeGraph
}

func (a *Archetype) Reset() {
	a.block.Reset()
	if a.graph != nil {
		a.graph.Reset()
	}
	a.Map.Reset()
	a.Mask = ArchetypeMask{}
	a.Id = NullArchetypeId
}

func (a *Archetype) InitArchetype(
	archId ArchetypeId,
	mask ArchetypeMask,
	colsInfos []ComponentInfo,
	initCapacity int,
) {
	a.Id = archId
	a.Mask = mask
	a.initCap = initCapacity
	a.graph = &ArchetypeGraph{} // Assuming ArchetypeGraph is defined elsewhere

	// 1. Initialize Column Map
	a.Map.Reset()

	// Entity ID is always at local index 0
	a.Map.Set(0, EntityColumnIndex)

	// Map components (Index 1..N)
	for i, info := range colsInfos {
		// i+1 because 0 is reserved for Entity
		a.Map.Set(info.ID, LocalColumnID(i+1))
	}

	// 2. Initialize Memory Block (Allocates memory)
	a.block.Init(initCapacity, colsInfos)
}

func (a *Archetype) Len() int {
	return int(a.block.Len)
}

func (a *Archetype) Cap() int {
	return int(a.block.Cap)
}

func (a *Archetype) GetEntityColumn() *Column {
	// Fast path: Entity is always at index 0
	return &a.block.Columns[EntityColumnIndex]
}

func (a *Archetype) GetColumn(id ComponentID) *Column {
	localIdx := a.Map.Get(id)
	if localIdx == InvalidLocalID {
		return nil
	}
	// Return pointer to the "Hot" column struct inside the block
	return &a.block.Columns[localIdx]
}

func (a *Archetype) registerEntity(entity Entity) ArchRow {
	// 1. Ensure space in the memory block
	a.block.EnsureCapacity(a.block.Len + 1)

	// 2. Write Entity ID to Column 0
	entityCol := &a.block.Columns[EntityColumnIndex]

	// Pointer arithmetic: Data + (Len * ItemSize)
	targetPtr := unsafe.Add(entityCol.Data, uintptr(a.block.Len)*entityCol.ItemSize)
	*(*Entity)(targetPtr) = entity

	// 3. Increment Row Count (Block manages length globally)
	newRow := ArchRow(a.block.Len)
	a.block.Len++

	return newRow
}

func (a *Archetype) SwapRemoveEntity(row ArchRow) (swapedEntity Entity, swaped bool) {
	lastRow := ArchRow(a.block.Len - 1)

	// Get the entity ID that is currently at the end (to return it)
	entityCol := &a.block.Columns[EntityColumnIndex]
	srcPtr := entityCol.GetElement(lastRow)
	entityToMove := *(*Entity)(srcPtr)

	// 1. Swap data in ALL active columns
	// We iterate over the block's columns directly, ignoring the Map (faster)
	for i := range a.block.Columns {
		col := &a.block.Columns[i]

		if row != lastRow {
			col.CopyData(row, lastRow)
		}

		// Optional: Zero out the memory at the old position to help debugging
		// and prevent stale pointers if strictly needed.
		col.ZeroData(lastRow)
	}

	// 2. Decrement length
	a.block.Len--

	if row == lastRow {
		return 0, false
	}
	return entityToMove, true
}

func (a *Archetype) linkNextArch(nextArch *Archetype, compID ComponentID) {
	a.graph.edgesNext[compID] = nextArch.Id
	nextArch.graph.edgesPrev[compID] = a.Id
}

func (a *Archetype) linkPrevArch(prevArch *Archetype, compID ComponentID) {
	a.graph.edgesPrev[compID] = prevArch.Id
	prevArch.graph.edgesNext[compID] = a.Id
}
