// Package colstore implements column-oriented storage for a single archetype.
//
// # Table
//
// [Table] is the central type. It embeds a [mem.Block] by value and adds
// column descriptors, a column index, and an [IDSeeder] hook:
//
//	Table
//	├── mem.Block      — growing collection of Chunks
//	├── []Column       — one entry per component column (entity column at index 0)
//	├── columnIndex    — maps comp.ID → columnPos in O(1)
//	└── IDSeeder       — fills entity IDs and registers them in the address index
//
// Each Chunk within the Table holds the entity ID column and one column per
// component type, laid out as parallel arrays at offsets stored in [Column]:
//
//	┌──────────────┬──────────────┬──────────────┬──────────────┐
//	│ []EntityID   │ []CompA      │ []CompB      │ []CompC ...  │
//	│ e0 e1 e2 …   │ a0 a1 a2 …   │ b0 b1 b2 …   │ c0 c1 c2 …   │
//	└──────────────┴──────────────┴──────────────┴──────────────┘
//	      ↑               ↑               ↑               ↑
//	  Column.Offset   Column.Offset   Column.Offset   Column.Offset
//
// Two entity allocation paths exist:
//   - [Table.AddEntity]       — allocates one slot and writes the entity ID.
//   - [Table.SpawnEntitySlice] — allocates n slots, invokes the [IDSeeder]
//     to fill entity IDs and register them, and returns a slice into the entity column.
//
// [Table.SwapRemoveEntity] removes the entity at a given [mem.BlockPos] using
// swap-and-pop, keeping all Chunks dense.
//
// # IDSeeder
//
// [IDSeeder] is a function type injected into each [Table] via [Table.SetIDSeeder].
// It decouples ID generation and address registration from the storage layer.
//
// # Column
//
// A [Column] describes one component column within a Chunk: its comp.ID,
// CompSize, and byte Offset within the Chunk. It exposes two accessors:
//   - At(chunk, slot) — pointer to a specific slot
//   - Base(chunk)     — pointer to the column's start within the Chunk
package colstore
