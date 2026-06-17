// Package ent owns the full lifecycle of ECS entities.
//
// An entity is an ID paired with a set of components:
//
//	entity = ID + components
//
// Component data lives in archetype tables (see [arch]), laid out as
// Structure-of-Arrays columns. From the outside an entity appears as a row
// spanning one column per component type it carries:
//
//	┌─────────────┬─────────────┬─────────────┬─────────────┐
//	│  EntityID   │    CompA    │    CompB    │    CompC    │
//	├─────────────┼─────────────┼─────────────┼─────────────┤
//	│     e0      │    a0       │    b0       │    c0       │
//	│     e1      │    a1       │    b1       │    c1       │
//	│     …       │    …        │    …        │    …        │
//	└─────────────┴─────────────┴─────────────┴─────────────┘
//
// Adding or removing a component moves the entity to a different archetype —
// the table whose column set matches the new component mask.
//
// # Manager
//
// [Manager] is the single owner of all entity state. It holds:
//
//   - [addr.Book]    — entity ID pool and entity-to-storage address index
//   - [arch.Catalog] — archetype graph and SoA tables
//
// On each new archetype, [Manager.Init] injects a [colstore.IDSeeder] closure
// into the archetype's table. The closure delegates to [addr.Book.Seed], so
// bulk entity spawning — ID allocation and address registration — happens
// entirely inside the storage layer.
//
// # Spawning
//
// Entities are spawned through [Factory], obtained via [Manager.CreateFactory].
// [Factory.Create] pre-allocates chunk slots; [Factory.Next] advances through
// each batch, populating [Factory.IDs] and [Factory.Cursor] for component writes.
//
// # Removal
//
// [Manager.Remove] looks up the entity's address via [addr.Book.Get], removes it
// from its archetype table with swap-and-pop (updating the displaced entity's
// address), clears the address entry, and recycles the ID.
//
// # Component Operations
//
// [Manager.UpsertComp] ensures the entity has the given component. If missing,
// it migrates the entity to the next archetype in the graph and returns a pointer
// to the fresh slot; if present, it returns a pointer to the existing data with
// no migration.
//
// [Manager.RemoveComp] migrates the entity to the archetype that excludes the
// given component. Graph edges are cached after the first traversal, so repeated
// structural changes follow a pre-resolved edge.
//
// [Manager.GetComp] resolves the entity's storage address via [addr.Book.Get]
// and returns an unsafe.Pointer to the component data in O(1).
package ent
