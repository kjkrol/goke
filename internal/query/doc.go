// Package query implements the query layer. It matches archetypes against
// component requirements and provides zero-allocation access to the matched
// data: a [Matcher] holds what a query matches, a [Walker] walks it.
//
// # Matcher
//
// A [Matcher] filters archetypes by component mask (include and exclude
// sets) and bakes each matching one into a table. Matchers are built once at
// initialization and updated automatically as new archetypes are created.
//
// # Walker
//
// A [Walker] walks one Matcher with a cursor of its own, in three patterns:
// All (chunk-by-chunk iteration), Pick (per-entity iteration over a given
// entity subset), and Seek (direct positioning on a single known entity,
// independent of the mask). Walkers of one Matcher may walk on as many
// goroutines while no entity or archetype changes.
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
// matcher whenever a new archetype is created. It keeps each Matcher's rule
// (masks and tracked columns) apart, so a new archetype is tried against the
// rules alone. It keeps them in blocks of a
// fixed size (InitialMatchers by default): a query holds a pointer to its
// Matcher, so a block never grows — a full one is followed by a new block —
// and no Matcher ever moves. Seek reads the tables baked as archetypes came,
// so walkers seeking on many goroutines only read their Matcher.
package query
