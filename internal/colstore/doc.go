// Package colstore implements column-oriented storage for a fixed set of
// component types.
//
// # Table
//
// [Table] is the central type. It manages a growing collection of fixed-size
// Chunks, each holding the entity ID column and one column per component type
// laid out as parallel arrays at offsets stored in [ColDef]:
//
//	┌──────────────┬──────────────┬──────────────┬──────────────┐
//	│ []EntityID   │ []CompA      │ []CompB      │ []CompC ...  │
//	│ e0 e1 e2 …   │ a0 a1 a2 …   │ b0 b1 b2 …   │ c0 c1 c2 …   │
//	└──────────────┴──────────────┴──────────────┴──────────────┘
//	      ↑               ↑               ↑               ↑
//	  ColDef.Offset   ColDef.Offset   ColDef.Offset   ColDef.Offset
//
// Two write paths exist:
//   - [Table.SpawnCursor]    — allocates n slots, invokes [IDSeeder] to fill entity IDs,
//     and populates the cursor for immediate component access. Used for initial entity creation.
//   - [Table.MoveEntityFrom] — allocates one slot, writes a known entity ID, copies matching
//     component columns from a source Table, then swap-removes the source slot. Used for migration.
//
// [Table.RemoveAt] removes the slot at a given [Pos] using swap-and-pop,
// keeping all Chunks dense.
//
// # IDSeeder
//
// [IDSeeder] is a function type injected into each [Table] via [Table.SetIDSeeder].
// It decouples ID generation and address registration from the storage layer.
//
// # ColDef
//
// A [ColDef] describes one component column within a Chunk: its comp.ID,
// CompSize, and byte Offset within the Chunk. It exposes two accessors:
//   - At(chunk, slot) — pointer to a specific slot
//   - Base(chunk)     — pointer to the column's start within the Chunk
package colstore
