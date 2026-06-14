// Package query implements the query layer. It matches archetypes against
// component requirements and provides zero-allocation iterators over the
// matched data.
//
// # View
//
// A [View] holds the component mask it requires (includeMask), the mask it
// rejects (excludeMask), and a Layout — the ordered list of component types
// the caller wants to read. It produces MatchedArchs — a slice of [MatchedArch]
// structs, one per matching archetype.
//
// Views are built once at initialization and kept up to date by
// [Registry].OnArchetypeCreated, which is called whenever a new archetype
// is registered via arch.Observer.
//
// # View Baking
//
// When a new archetype matches a view's masks, [View.AddArchetype] is called.
// It reads each [colstore.Column]'s PageOffset and ItemSize fields and caches:
//   - EntityPageOffset  — byte offset of the entity column within a Chunk
//   - CompOffsets[i]    — byte offset of each requested component column
//   - CompSizes[i]      — item size of each requested component
//
// These values are used directly at iteration time, so the hot path is a
// series of pointer additions with no column lookup.
//
// # MatchedArch
//
// [MatchedArch] carries a pointer to the matched [colstore.Table] and the
// precomputed offsets and sizes for the view's layout. At iteration time the
// hot path reads Table.Chunks directly and applies CompOffsets/CompSizes via
// pointer arithmetic — no column lookup, no hash map.
//
// # archMapping
//
// A flat []int32 indexed by arch.ID. Maps an archetype ID to its position in
// MatchedArchs, or -1 if not matched. Used by [View.GetMatchedArch] in the
// Filter path to resolve per-entity archetype membership in O(1).
//
// # Registry
//
// [Registry] holds all registered Views and implements arch.Observer.
// On OnArchetypeCreated it fans out to every view whose mask matches the new
// archetype, keeping all queries current without a full reindex.
package query
