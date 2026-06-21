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
// GetComp, CreateFactory.
//
// [Factory] handles bulk entity creation using a chunk-based iterator.
package ent
