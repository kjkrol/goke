// Package reg implements the top-level world registry — a columnar component
// storage organized by archetype. Instead of storing records as individual
// objects, it groups entities by their component composition (archetype) and
// lays out each component type as a contiguous Structure-of-Arrays column,
// enabling cache-friendly bulk iteration and efficient structural mutations.
//
// It can be used as the storage layer of an Entity Component System.
//
// # Entity Lifecycle
//
// [Registry] assigns each entity a unique generational identifier ([uid.UID64])
// and tracks its current archetype and storage position ([soa.BlockPos]).
// CreateEntity places a new entity in the root archetype (no components).
// RemoveEntity unlinks it via swap-and-pop and recycles its identifier.
//
// # Component Operations
//
// [Registry.UpsertComp] is an upsert: if the entity does not yet have the
// component it migrates to a new archetype and returns a pointer to the fresh
// slot; if the component already exists it returns a pointer to the existing
// data with no migration. Callers can treat it as "ensure and access" without
// a prior existence check.
//
// [Registry.RemoveComp] removes a component from an entity, migrating it to
// the archetype that excludes that component. The archetype graph caches these
// transitions so repeated structural changes follow a pre-resolved edge rather
// than a map lookup.
//
// [Registry.RegCompType] maps a Go reflect.Type to a [comp.Meta] descriptor,
// assigning a stable [comp.ID] used by all subsequent operations.
//
// [Registry.GetComp] returns an unsafe.Pointer to an entity's component data
// in O(1) via [arch.EntityIndex], with no hash map involved.
//
// # Dense Packing
//
// Entities are stored in fixed-size chunks sized to fit within L1 data cache.
// Swap-and-pop removal keeps chunks dense — no holes, no background
// defragmentation. The result is that iterating over a component type reads
// a compact, cache-warm sequence of values regardless of which entities are alive.
//
// # Registry
//
// [Registry] wires together four subsystems:
//
//   - [comp.Catalog]   — maps Go types to [comp.Meta]
//   - [arch.Catalog]   — manages archetypes, entity placement, and graph edges
//   - [query.Registry] — keeps active queries current as archetypes are created
//   - [uid.UID64Pool]  — issues and recycles generational entity identifiers
package reg
