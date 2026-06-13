// Package soa implements the paged Structure-of-Arrays (SoA) memory primitives
// used by archetypes to store component data.
//
// # Chunk
//
// A [Chunk] is a single fixed-size memory slab — one contiguous []byte allocation
// whose layout is determined by [ChunkLayout]. All component columns for one
// archetype live inside the same Chunk, interleaved at precomputed offsets.
//
//	Chunk (single allocation)
//
//	┌────────────────────────────────────────────────────────────┐
//	│ data []byte                                                │
//	├────────────────────────────────────────────────────────────┤
//	│ Entity Column │ CompA Column │ CompB Column │ ...          │
//	└────────────────────────────────────────────────────────────┘
//
// From the caller's perspective each column is a typed slice of N slots:
//
//	┌─────────────┬─────────────┬─────────────┬─────────────┐
//	│ []Entity    │ []CompA     │ []CompB     │ []CompC     │
//	├─────────────┼─────────────┼─────────────┼─────────────┤
//	│ e0  a0  b0  │ e1  a1  b1  │ ...                        │
//	└─────────────┴─────────────┴─────────────┴─────────────┘
//
// # ChunkLayout and CalculateLayout
//
// [CalculateLayout] determines how many entities fit in a single Chunk such that
// the total allocation stays within [L1DataCacheSize]. It aligns each column to
// its component's natural alignment and iterates capacity downward until the
// layout fits.
//
// The resulting [ChunkLayout] holds:
//   - ChunkCap   — max entities per Chunk
//   - ChunkBytes — total byte size of one Chunk
//   - Offsets    — byte offset of each column within the Chunk
//
// # Block
//
// A [Block] is a dynamically growing collection of Chunks sharing one [ChunkLayout].
// It allocates new Chunks on demand and tracks total entity count across all Chunks.
// [Block.AllocSlot] returns a [BlockPos] identifying the exact Chunk and slot
// assigned to a new entity.
//
// # BlockPos
//
// A [BlockPos] identifies a slot within a Block: ChunkIdx (which Chunk) and
// ChunkSlot (which slot within that Chunk). It replaces the raw pair
// (ChunkIdx, ChunkSlot) wherever both values must travel together.
//
// # L1DataCacheSize
//
// Defined in const_default.go / const_arm64.go via build tags.
// Used by [CalculateLayout] to size Chunks so that a full Chunk fits within
// the L1 data cache, maximising iteration cache locality.
package soa
