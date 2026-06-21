// Package query implements the query layer. It matches archetypes against
// component requirements and provides zero-allocation iterators over the
// matched data.
//
// # View
//
// A [View] filters archetypes by component mask (include and exclude sets)
// and exposes two iteration modes: All (chunk-by-chunk) and Filter (per-entity).
// Views are built once at initialization and updated automatically as new
// archetypes are created.
//
// # BakedTable
//
// For each matching archetype, a [BakedTable] stores a pointer to the
// archetype's [colstore.Table] alongside precomputed per-column byte offsets.
// At iteration time the hot path is pure pointer arithmetic — no column
// lookup, no hash map.
//
// # Catalog
//
// [Catalog] holds all registered Views and fans out to each matching view
// whenever a new archetype is created.
package query
