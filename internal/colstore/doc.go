// Package colstore implements column-oriented storage for a single archetype.
//
// # Table
//
// [Table] is the central type. It embeds a [soa.Block] by value (zero extra
// pointer indirection) and adds a column descriptor slice and a column index:
//
//	Table
//	├── soa.Block          — Chunks + ChunkLayout + Len + Reserved
//	├── []Column           — one entry per component column (entity column at index 0)
//	└── columnIndex        — maps comp.ID → columnPos in O(1)
//
// [Table.AddEntity] allocates a slot via [soa.Block.AllocSlot] and writes the
// entity ID into the entity column, returning the resulting [soa.BlockPos].
// [Table.SwapRemoveEntity] removes the entity at a given [soa.BlockPos] using
// swap-and-pop, keeping all Chunks dense.
//
// # Column
//
// A [Column] describes one component column within a Chunk: its comp.ID,
// CompSize, and byte offset within the Chunk (Offset). It exposes two
// accessors:
//   - At(chunk, slot) — pointer to a specific slot
//   - Base(chunk)     — pointer to the column's start within the Chunk
//
// # columnIndex
//
// An unexported fixed-size array mapping each [comp.ID] to a columnPos —
// the local index of that component's [Column] within Table.columns.
// Lookups are a single array read: O(1) with no hashing or scanning.
package colstore
