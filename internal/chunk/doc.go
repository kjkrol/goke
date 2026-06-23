// Package chunk implements cache-aligned chunked memory blocks using a
// Structure of Arrays (SoA) layout. Unlike the conventional Array of
// Structures (AoS) approach, SoA stores each field in its own contiguous
// array, so iterating over a single field accesses a linear memory region —
// fewer cache misses, better prefetcher utilisation, and SIMD-friendly access
// patterns.
//
// # Layout
//
// [Layout.Init] computes how many slots of each field type fit within
// [L1DataCacheSize]. Fields are arranged as parallel arrays at precomputed
// byte offsets:
//
//	┌───────────────────────────────────────┐
//	│             one chunk                 │
//	├──────────┬──────────┬──────────┬──────┤
//	│ []type0  │ []type1  │ []type2  │ ...  │
//	└──────────┴──────────┴──────────┴──────┘
//	     ↑          ↑          ↑
//	  Offsets[0] Offsets[1] Offsets[2]
//
// The resulting [Layout] holds:
//   - ChunkCap   — max slots per chunk
//   - ChunkBytes — total byte size of one chunk
//   - Offsets    — byte offset of each field array within the chunk
//
// # Pack
//
// [Pack] is a densely packed, dynamically growing sequence of fixed-size chunks
// sharing one [Layout]. New chunks are allocated as a single contiguous []byte
// to keep backing memory cache-friendly. [Pack.ResolveTail] trims empty
// trailing chunks as the Pack shrinks, but keeps the backing array of the
// last one trimmed as a spare — [Pack.AddChunks] reuses it on the next
// growth instead of allocating fresh memory, which makes repeated
// shrink/regrow cycles (e.g. structural edits that migrate entities back and
// forth between two archetypes) effectively allocation-free after the first
// cycle. [Pack.Purge] releases that spare (and any other trimmable chunks)
// immediately, for callers that know a Pack won't be repopulated soon.
//
// # Pos
//
// [Pos] identifies a slot within a Pack: [Idx] (which chunk) and
// [Slot] (which slot within that chunk).
//
// # L1DataCacheSize
//
// Defined per platform via build tags. [Layout.Init] uses it to size chunks
// so that a full chunk fits within the L1 data cache.
package chunk
