// Package migration implements bulk archetype migration for batches of
// entities, sharing one hard contract between [Migrator] and [Remover]:
//
// Callers must pass ids in chunk-iteration order — one call per chunk, ids
// as returned by Query.Next()/Cursor().IDs. Every efficiency gain below
// depends on it; passing ids in arbitrary order silently degrades both
// types to per-entity cost with no error raised.
//
// That contract buys three things, in resolveSlotRefs and
// Defragmenter.Compact (both shared code, used by Migrator and Remover
// alike):
//
//  1. The SlotAligned fast path skips addr.Book entirely — it assumes the
//     batch is the *whole* unchanged chunk and synthesizes slot = 0..n-1.
//  2. The slow path (source table changed since snapshot) still amortizes:
//     consecutive ids sharing a chunk reuse one ChunkIdxByPtr lookup instead
//     of paying it per entity.
//  3. Compact receives slot-ordered holes, so it can detect contiguous runs
//     and take modeAllMigrate/modeChunkSwap instead of the per-slot
//     modeSlotLevel fallback.
//
// # Migrator
//
// Applies a fixed set of component adds/removes to a batch sharing one
// source archetype. The destination archetype is resolved once per source
// (O(1) array read after that) and memoized — computed directly from masks,
// so it never creates intermediate edge-graph archetypes.
//
// # Remover
//
// Migrator's removal-only counterpart: no destination archetype, so no
// bytes move. It reuses Migrator's id resolution and compaction step
// (resolveSlotRefs, removeBatch) — the same code that backs Migrator's own
// full-unlink case (every component removed).
package migration
