// Package arch implements the archetype graph, entity-to-storage mapping,
// and entity lifecycle operations. It also defines the archetype identity
// type ([ID]) and its associated data structures.
//
// # Archetype-Based Storage
//
// Entities are grouped into archetypes according to their component composition.
// Each archetype is identified by a [comp.Mask] and owns a [colstore.Table]
// — a [soa.Block] of [soa.Chunk]s with one column per component type.
//
// Adding or removing a component moves an entity to a different archetype.
// The archetype graph caches these transitions so that repeated structural
// changes on the same composition follow a pre-resolved edge rather than
// a map lookup.
//
// # Graph
//
// Each [Archetype] owns a [Graph] — two fixed-size arrays of [ID] edges:
//
//   - edgesNext[compID] — the archetype reached by adding component compID
//   - edgesPrev[compID] — the archetype reached by removing component compID
//
// On the first transition for a given composition, [Catalog] resolves
// (or creates) the target archetype and writes the edge. Subsequent transitions
// follow the cached edge directly — a single array lookup.
//
// # EntityIndex
//
// A flat slice indexed by entity index (extracted from uid.UID64).
// Each slot holds an [EntityLocation]:
//
//	EntityID
//	    ↓
//	EntityIndex[index]
//	    ↓
//	{ ArchId, Pos, Generation }
//
// Pos is a [soa.BlockPos] — the ChunkIdx and ChunkSlot within the archetype's
// [colstore.Table] — providing O(1) access to an entity's storage location
// without hash maps or archetype scans. The Generation field validates that a
// looked-up slot still belongs to the expected entity, preventing stale-reference
// access after recycling.
//
// # Dense Storage (Swap-and-Pop)
//
// When an entity is removed, its slot is filled by moving the last entity in
// the same chunk into the freed position. This swap-and-pop strategy avoids
// holes in memory, maintains dense storage, and eliminates background
// defragmentation. The displaced entity's EntityIndex entry is updated
// to reflect its new position.
//
// # Observer
//
// [Catalog] calls [Observer].OnArchetypeCreated each time a new archetype is
// registered. query.Registry implements this interface to update active queries
// without creating a circular import (arch → query → arch). This notification
// is on the cold path and carries no performance cost during normal iteration.
package arch
