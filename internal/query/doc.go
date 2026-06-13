// Package query implements the query layer. It matches archetypes against
// component requirements and provides zero-allocation iterators over the
// matched data.
//
// # View
//
// A [View] holds the component mask it requires (includeMask), the mask it
// rejects (excludeMask), and a Layout — the ordered list of component types
// the caller wants to read. It produces BakedTables — a slice of [BakedTable]
// structs, one per matching archetype.
//
// Views are built once at initialization and kept up to date by
// [Catalog.OnArchetypeCreated], which is passed as a callback to arch.Catalog
// and called whenever a new archetype is registered.
//
// # View Baking
//
// When a new archetype matches a view's masks, [View.Bake] is called.
// It reads each [colstore.Column]'s Offset and CompSize fields and caches:
//   - CompOffsets[i]    — byte offset of each requested component column
//   - CompSizes[i]      — item size of each requested component
//
// These values are used directly at iteration time, so the hot path is a
// series of pointer additions with no column lookup.
//
// # BakedTable
//
// [BakedTable] carries a pointer to the matched [colstore.Table] and the
// precomputed offsets and sizes for the view's layout. At iteration time the
// hot path reads Table.Chunks directly and applies CompOffsets/CompSizes via
// pointer arithmetic — no column lookup, no hash map.
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
