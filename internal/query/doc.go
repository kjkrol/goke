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
// SeekH is Seek's per-entity fast path: it skips the archetype-change and
// alive checks, trusting the caller to have already established the target
// archetype via a prior Seek on the same entity batch.
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
// matcher whenever a new archetype is created. It keeps them in blocks of a
// fixed size (InitialMatchers by default): a query holds a pointer to its
// Matcher, so a block never grows — a full one is followed by a new block —
// and no Matcher ever moves. A Matcher's Seek bakes grow with the archetypes
// it meets, so it stays small whatever the archetype limit.
package query
