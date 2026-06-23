// Package query implements the query layer. It matches archetypes against
// component requirements and provides zero-allocation access to the matched
// data through a single type, [Matcher], with three access patterns.
//
// # Matcher
//
// A [Matcher] filters archetypes by component mask (include and exclude
// sets) and exposes three access patterns: All (chunk-by-chunk iteration),
// Pick (per-entity iteration over a given entity subset), and Seek (direct
// positioning on a single known entity, independent of the mask). Matchers
// are built once at initialization and updated automatically as new
// archetypes are created.
//
// # BakedTable
//
// For each matching archetype, a [BakedTable] stores a pointer to the
// archetype's column table alongside precomputed per-column byte offsets.
// At iteration time the hot path is pure pointer arithmetic — no column
// lookup, no hash map.
//
// # Catalog
//
// [Catalog] holds all registered Matchers and fans out to each matching
// matcher whenever a new archetype is created.
package query
