// Package migration implements bulk archetype migration for batches of entities.
//
// It is the batch counterpart of [ent.Editor]: where Editor migrates one entity
// at a time with immediate swap-and-pop compaction, [Migrator] defers compaction
// until the full batch is known, enabling block-level column copies and direct
// per-entity address-index writes with no intermediate buffering.
//
// # Migrator
//
// [Migrator] applies a fixed set of structural changes — adding and removing
// components — to a batch of entities that all share the same source archetype.
// The destination archetype is pre-computed eagerly for every source archetype
// when it is added to the catalog, so no graph traversal happens on the hot path.
//
//  1. Look up source → destination archetype (O(1) array read).
//  2. Copy component data in contiguous runs from source columns to destination
//     columns, updating the address book as entities land.
//  3. Compact the source archetype knowing all hole positions upfront, updating
//     the address book for relocated survivors.
package migration
