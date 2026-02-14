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
// Column Descriptor
// -----------------------------------------------------------------------------

type Column struct {
	CompID      ComponentID
	ItemSize    uintptr
	ChunkOffset uintptr // Offset from the start of the chunk data to this column's start
}

// GetPointer returns the unsafe pointer to the specific element in the given chunk.
// Formula: Chunk.Data + ColumnOffset + (Row * ItemSize)
// Cost: Simple pointer arithmetic, very fast.
func (c *Column) GetPointer(chunk *chunk, row ChunkRow) unsafe.Pointer {
	// Base address of the column within this chunk
	colBase := unsafe.Add(chunk.ptr, c.ChunkOffset)
	// Address of the specific row
	return unsafe.Add(colBase, uintptr(row)*c.ItemSize)
}

// -----------------------------------------------------------------------------
// Archetype
// -----------------------------------------------------------------------------

type Archetype struct {
	Mask    ArchetypeMask
	Id      ArchetypeId
	Map     ColumnMap
	Memory  Memo
	Columns []Column
	graph   *ArchetypeGraph
}

func (a *Archetype) Reset() {
	a.Memory.Clear()
	if a.graph != nil {
		a.graph.Reset()
	}
	a.Map.Reset()
	a.Mask = ArchetypeMask{}
	a.Id = NullArchetypeId
	a.Columns = nil
}

func (a *Archetype) InitArchetype(
	archId ArchetypeId,
	mask ArchetypeMask,
	colsInfos []ComponentInfo,
) {
	a.Id = archId
	a.Mask = mask
	a.graph = &ArchetypeGraph{} // Assuming ArchetypeGraph is defined elsewhere

	// 1. Initialize Memory (Calculate Layout)
	// This computes how components are packed into 16KB chunks.
	a.Memory.Init(colsInfos)

	// 2. Setup Columns & Map
	// We need space for Entity Column + Data Columns
	count := len(colsInfos) + 1
	a.Columns = make([]Column, count)
	a.Map.Reset()

	// --- A. Setup Entity Column (LocalID 0) ---
	a.Map.Set(0, EntityColumnIndex)
	a.Columns[EntityColumnIndex] = Column{
		CompID:      0, //TODO: this is incorrect
		ItemSize:    unsafe.Sizeof(Entity(0)),
		ChunkOffset: a.Memory.Layout.Offsets[0], // Offset from Layout calculation
	}

	// --- B. Setup Component Columns (LocalID 1..N) ---
	for i, info := range colsInfos {
		localIdx := LocalColumnID(i + 1)

		a.Map.Set(info.ID, localIdx)

		a.Columns[localIdx] = Column{
			CompID:      info.ID,
			ItemSize:    info.Size,
			ChunkOffset: a.Memory.Layout.Offsets[i+1], // +1 because index 0 is Entity
		}
	}
}

// Len returns the total number of entities in this archetype.
func (a *Archetype) Len() int {
	return int(a.Memory.Len)
}

// GetEntityColumn returns the accessor for the Entity ID column.
func (a *Archetype) GetEntityColumn() *Column {
	// Fast path: Entity is always at index 0
	return &a.Columns[EntityColumnIndex]
}

// GetColumn returns the accessor for a specific component.
func (a *Archetype) GetColumn(id ComponentID) *Column {
	localIdx := a.Map.Get(id)
	if localIdx == InvalidLocalID {
		return nil
	}
	// Return pointer to the "Hot" column struct inside the block
	return &a.Columns[localIdx]
}

// AddEntity reserves a slot and writes the Entity ID.
func (a *Archetype) AddEntity(entity Entity) (ChunkIdx, ChunkRow) {
	chunk, chunkIdx, row := a.Memory.AllocSlot()

	entityCol := &a.Columns[EntityColumnIndex]
	destPtr := entityCol.GetPointer(chunk, row)
	*(*Entity)(destPtr) = entity

	return chunkIdx, row
}

// SwapRemoveEntity removes an entity at the specified location (O(1)).
// It moves the last entity of the archetype into the empty slot (Swap).
// Returns the entity that was moved (swappedEntity) and true if a move happened.
func (a *Archetype) SwapRemoveEntity(targetChunkIdx ChunkIdx, targetRow ChunkRow) (swappedEntity Entity, swapped bool) {
	// 1. Identify the Tail (The last entity in the last chunk)
	lastChunkIdx := ChunkIdx(len(a.Memory.Pages) - 1)
	lastChunk := a.Memory.Pages[lastChunkIdx]

	// FIX: Jeśli ostatni chunk jest pusty i nie jest to jedyny chunk,
	// musimy go usunąć ze slice'a, aby ogon wskazywał na chunk z danymi.
	for lastChunk.Len == 0 && lastChunkIdx > 0 {
		a.Memory.Pages = a.Memory.Pages[:lastChunkIdx] // Usuń pusty chunk z końca
		lastChunkIdx--
		lastChunk = a.Memory.Pages[lastChunkIdx]
	}

	lastRow := lastChunk.Len - 1

	// Identify the Target Chunk (where the hole is)
	// We access the slice directly using the index from LinkStore.
	targetChunk := a.Memory.Pages[targetChunkIdx]

	// Edge Case: Removing the very last entity (No swap needed)
	if targetChunkIdx == lastChunkIdx && targetRow == lastRow {
		a.zeroEntityAt(lastChunk, lastRow)
		lastChunk.Len--
		a.Memory.Len--
		return 0, false
	}

	// 2. Retrieve the Entity ID from the Tail (this is the one moving)
	entityCol := &a.Columns[EntityColumnIndex]
	srcEntityPtr := entityCol.GetPointer(lastChunk, lastRow)
	entityToMove := *(*Entity)(srcEntityPtr)

	// 3. Move Data: Tail -> Hole (Target)
	// We iterate over all columns (Entity + Components)
	for i := range a.Columns {
		col := &a.Columns[i]

		src := col.GetPointer(lastChunk, lastRow)
		dst := col.GetPointer(targetChunk, targetRow)

		// Memcpy
		// Go's internal copy handles overlap safely, though here memory is distinct.
		// We cast to byte slice for copying.
		srcSlice := unsafe.Slice((*byte)(src), col.ItemSize)
		dstSlice := unsafe.Slice((*byte)(dst), col.ItemSize)
		copy(dstSlice, srcSlice)
	}

	// 4. Cleanup the old Tail position (GC safety)
	a.zeroEntityAt(lastChunk, lastRow)

	// 5. Update Counters
	lastChunk.Len--
	a.Memory.Len--

	return entityToMove, true
}

// zeroEntityAt clears memory at the given location to prevent stale pointers (GC).
func (a *Archetype) zeroEntityAt(c *chunk, row ChunkRow) {
	for i := range a.Columns {
		col := &a.Columns[i]
		ptr := col.GetPointer(c, row)

		// Efficient memory zeroing
		b := unsafe.Slice((*byte)(ptr), col.ItemSize)
		clear(b)
	}
}

// -----------------------------------------------------------------------------
// Graph Linking
// -----------------------------------------------------------------------------

func (a *Archetype) linkNextArch(nextArch *Archetype, compID ComponentID) {
	a.graph.edgesNext[compID] = nextArch.Id
	nextArch.graph.edgesPrev[compID] = a.Id
}

func (a *Archetype) linkPrevArch(prevArch *Archetype, compID ComponentID) {
	a.graph.edgesPrev[compID] = prevArch.Id
	prevArch.graph.edgesNext[compID] = a.Id
}
