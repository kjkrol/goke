// Package query implements the query layer. It matches archetypes against
// component requirements and provides zero-allocation iterators over the
// matched data.
//
// # View
//
// A [View] holds the component mask it requires (includeMask), the mask it
// rejects (excludeMask), and a [comp.Composition] — the ordered list of
// component types the caller wants to read. It produces BakedTables — a slice
// of [BakedTable] structs, one per matching archetype.
//
// Views are built once at initialization and kept up to date by
// [Catalog.OnArchetypeCreated], registered as a callback on the archetype
// catalog and called whenever a new archetype is created.
//
// # View Baking
//
// When a new archetype matches a view's masks, [View.BakeIfMatch] is called.
// It reads each [colstore.Column]'s Offset field and caches:
//   - CompOffsets[i]    — byte offset of each requested component column
//
// These values are used directly at iteration time, so the hot path is a
// series of pointer additions with no column lookup. The per-row stride comes
// from unsafe.Sizeof(T) at each typed access site (a compile-time constant),
// so no runtime item-size table is needed.
//
// # BakedTable
//
// [BakedTable] carries a pointer to the matched [colstore.Table] and the
// precomputed column offsets for the view's composition. At iteration time the
// hot path advances chunk by chunk and applies CompOffsets via pointer
// arithmetic — no column lookup, no hash map.
//
// # BakedTablesCatalog
//
// Owned by each [View] via embedding. Holds the []BakedTable slice and a flat
// []int32 archTableIndex indexed by arch.ID. Maps an archetype ID to its position
// in BakedTables (-1 if not matched), enabling O(1) lookup via [BakedTablesCatalog.Get]
// in the Filter path.
//
// # Catalog
//
// [Catalog] holds all registered Views. Its OnArchetypeCreated method is passed
// as a callback to arch.Catalog and fans out to every view whose mask matches
// the new archetype, keeping all queries current without a full reindex.
package query
