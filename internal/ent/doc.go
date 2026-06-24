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
// # Editor
//
// [Editor] applies a fixed set of component adds/removes to an entity as a
// single archetype migration. Its cost scales with the width of the source
// and destination archetypes, not with how many components the edit
// changes — removing a few components from a wide archetype costs about as
// much as removing many. When the same Editor repeatedly migrates entities
// back and forth between the same two archetypes (e.g. toggling a component
// every tick), the underlying storage reuses the freed capacity instead of
// reallocating it, making the cycle allocation-free after the first one.
package ent
