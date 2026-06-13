// Package entity owns the full lifecycle of ECS entities.
//
// An entity is nothing more than an ID paired with a set of components:
//
//	entity = ID + components
//
// The components themselves live in archetype tables (see [arch]), laid out as
// Structure-of-Arrays columns. From the outside an entity looks like a row that
// spans one column per component type it carries:
//
//	┌─────────────┬─────────────┬─────────────┬─────────────┐
//	│  EntityID   │    CompA    │    CompB    │    CompC    │
//	├─────────────┼─────────────┼─────────────┼─────────────┤
//	│     e0      │    a0       │    b0       │    c0       │
//	│     e1      │    a1       │    b1       │    c1       │
//	│     …       │    …        │    …        │    …        │
//	└─────────────┴─────────────┴─────────────┴─────────────┘
//
// Adding or removing a component moves the entity to a different archetype
// (a different table whose column set matches the new component mask).
//
// # EntityLocation
//
// [EntityLocation] records where an entity currently lives in storage:
//
//   - ArchId      — which archetype table the entity belongs to
//   - Pos         — the [soa.BlockPos] (ChunkIdx + ChunkSlot) within that table
//   - Generation  — guards against stale access after an ID is recycled
//
// # Index
//
// [Index] is a flat slice keyed by the entity's numeric index (low bits of [uid.UID64]).
// It maps an entity ID to its [EntityLocation] in O(1) — no hash map, no scanning.
//
//	uid.UID64 → Unpack() → (index, generation)
//	                            │
//	                   Index.entries[index]
//	                            │
//	                   EntityLocation { ArchId, Pos, Generation }
//
// On every read the stored Generation is compared against the requested one.
// A mismatch means the slot was recycled and the lookup returns false.
//
// # Manager
//
// [Manager] is the single owner of all entity state. It combines:
//
//   - uid.UID64Pool  — allocates and recycles entity IDs
//   - [Index]        — tracks each entity's current storage position
//   - [arch.Catalog] — owns the archetype graph and SoA tables
//
// # Entity Lifecycle
//
// [Manager.Create] assigns a new entity a unique generational identifier ([uid.UID64])
// and places it in the root archetype (no components).
// [Manager.Remove] is the reverse: swap-and-pop removal from the archetype table,
// index clear, and ID recycle.
//
// # Component Operations
//
// [Manager.UpsertComp] is an upsert: if the entity does not yet have the component
// it migrates to a new archetype and returns a pointer to the fresh slot; if the
// component already exists it returns a pointer to the existing data with no
// migration. Callers can treat it as "ensure and access" without a prior existence check.
//
// [Manager.RemoveComp] removes a component from an entity, migrating it to the
// archetype that excludes that component. The archetype graph caches these transitions
// so repeated structural changes follow a pre-resolved edge rather than a map lookup.
//
// [Manager.GetComp] returns an unsafe.Pointer to the entity's component data
// in O(1) via [Index], with no hash map involved.
//
// # Dense Packing
//
// Entities are stored in fixed-size chunks sized to fit within L1 data cache.
// Swap-and-pop removal keeps chunks dense — no holes, no background defragmentation.
// The result is that iterating over a component type reads a compact, cache-warm
// sequence of values regardless of which entities are alive.
//
// # Blueprint Fast Path
//
// Bulk creation (blueprints) uses [Manager.NextID] + [Manager.UpdateLocation]
// to bypass the single-entity Create path and write entities directly into a
// pre-resolved archetype table.
package ent
