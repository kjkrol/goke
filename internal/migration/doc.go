// Package migration implements bulk archetype migration for batches of
// entities: [Migrator], [Remover], and [ValueMigrator].
//
// Callers must pass ids in chunk-iteration order — one call per chunk, ids
// as returned by Query.Next()/Cursor().IDs. Passing ids in arbitrary order
// silently degrades performance; it never breaks correctness. Three cases
// determine that cost:
//
//  1. Nothing changed since capture: ids are placed with no per-entity
//     lookups at all.
//  2. Something changed since capture: each id is re-validated, but
//     consecutive ids sharing a chunk still amortize the cost.
//  3. Vacated slots left behind are compacted in one pass — contiguous runs
//     move as a block rather than slot by slot.
//
// # Migrator
//
// Applies a fixed set of component adds/removes to a batch sharing one
// source archetype.
//
// # Remover
//
// Removes every entity in a batch sharing one source archetype outright.
//
// # ValueMigrator
//
// Adds one component to a batch sharing one source archetype and writes a
// caller-supplied value into it per entity.
package migration
