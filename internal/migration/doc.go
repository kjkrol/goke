// Package migration implements bulk archetype migration for batches of entities.
//
// It is the batch counterpart of [ent.Editor]: where Editor migrates one entity
// at a time with immediate swap-and-pop compaction, [Migrator] defers compaction
// until the full batch is known, enabling block-level column copies and a
// single homogeneous pass of address-index updates per batch.
//
// # Migrator
//
// [Migrator] applies a fixed set of structural changes — adding and removing
// components — to a batch of entities that all share the same source archetype.
// The destination archetype is resolved lazily on the first migration from a
// given source and memoized in a flat array: the target composition is computed
// directly from masks, creating no intermediate edge-graph archetypes, so
// coexisting migrators cannot cross-multiply the archetype catalog.
//
//  1. Look up source → destination archetype (O(1) array read, resolved once).
//  2. Move bytes: copy component data in contiguous runs from source columns
//     to destination columns, then compact the source archetype knowing all
//     hole positions upfront.
//  3. Update the address book in one pass: destination positions recomputed
//     arithmetically from per-chunk geometry, relocated survivors from the
//     compaction's move list.
package migration
