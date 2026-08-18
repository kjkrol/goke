// Package ent is the internal API layer for entity lifecycle.
// All entity operations — creation, destruction, and component migration —
// flow through this package.
//
// An entity is an ID paired with a set of components. Internally it occupies
// a row across the SoA columns of its archetype table:
//
//	┌─────────────┬─────────────┬─────────────┬─────────────┐
//	│  EntityID   │    CompA    │    CompB    │    CompC    │
//	├─────────────┼─────────────┼─────────────┼─────────────┤
//	│     e0      │    a0       │    b0       │    c0       │
//	│     e1      │    a1       │    b1       │    c1       │
//	│     …       │    …        │    …        │    …        │
//	└─────────────┴─────────────┴─────────────┴─────────────┘
//
// Adding or removing a component moves the entity to a different archetype.
//
// [Manager] delegates storage to [arch.Catalog] and identity management
// to [addr.Book], exposing a unified API: Remove, UpsertComp, RemoveComp,
// CreateFactory.
//
// [Factory] handles bulk entity creation using a chunk-based iterator.
//
// [Editor], [Remover], and [ValueEditor] apply bulk archetype migrations to
// batches of entities sharing one source archetype. Callers must pass ids in
// chunk-iteration order — one call per chunk, ids as returned by
// Query.Next()/Cursor().IDs. Passing ids in arbitrary order silently
// degrades performance; it never breaks correctness. Three cases determine
// that cost:
//
//  1. Nothing changed since capture: ids are placed with no per-entity
//     lookups at all.
//  2. Something changed since capture: each id is re-validated, but
//     consecutive ids sharing a chunk still amortize the cost.
//  3. Vacated slots left behind are compacted in one pass — contiguous runs
//     move as a block rather than slot by slot.
//
// # Editor
//
// Applies a fixed set of component adds/removes to a batch sharing one
// source archetype.
//
// # Remover
//
// Removes every entity in a batch sharing one source archetype outright.
//
// # ValueEditor
//
// Adds one component to a batch sharing one source archetype and writes a
// caller-supplied value into it per entity.
package ent
