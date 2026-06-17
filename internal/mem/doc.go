// Package mem implements cache-aligned chunked memory blocks.
//
// # ChunkLayout
//
// [ChunkLayout.Init] computes how many slots of each field type fit within
// [L1DataCacheSize]. Fields are arranged as parallel arrays at precomputed
// byte offsets:
//
//	┌───────────────────────────────────────┐
//	│             one Chunk                 │
//	├──────────┬──────────┬──────────┬──────┤
//	│ []type0  │ []type1  │ []type2  │ ...  │
//	└──────────┴──────────┴──────────┴──────┘
//	     ↑          ↑          ↑
//	  Offsets[0] Offsets[1] Offsets[2]
//
// The resulting [ChunkLayout] holds:
//   - ChunkCap   — max slots per Chunk
//   - ChunkBytes — total byte size of one Chunk
//   - Offsets    — byte offset of each field array within the Chunk
//
// # Block
//
// [Block] is a dynamically growing collection of fixed-size Chunks sharing
// one [ChunkLayout]. New Chunks are allocated as a single contiguous []byte
// to keep backing memory cache-friendly.
//
// Chunk growth:
//   - [Block.AddChunks]    — allocates n new Chunks as one contiguous []byte.
//
// Slot allocation:
//   - [Block.AllocSlot]    — appends one slot and returns the resulting [BlockPos].
//   - [Block.AllocSlots]   — advances a Chunk's fill mark by n slots at once.
//   - [Block.PrepareSlots] — ensures enough Chunks exist for count slots starting
//     from the current tail; returns the starting [ChunkIdx] and the number of slots
//     available in that first Chunk.
//
// # BlockPos
//
// [BlockPos] identifies a slot within a Block: [ChunkIdx] (which Chunk) and
// [ChunkSlot] (which slot within that Chunk).
//
// # L1DataCacheSize
//
// Defined per platform via build tags. [ChunkLayout.Init] uses it to size Chunks
// so that a full Chunk fits within the L1 data cache.
package mem
