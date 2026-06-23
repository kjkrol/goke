// Package arch implements the archetype graph and archetype-based SoA storage.
// It manages which component compositions exist and caches the structural transitions
// between them as graph edges.
//
// # Archetype-Based Storage
//
// Entities are grouped into archetypes according to their component composition.
// Each archetype is identified by a [comp.Mask] and owns a [colstore.Table]
// — a column-oriented store of fixed-size Chunks with one column per component type.
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
// # Dense Storage (Swap-and-Pop)
//
// When an entity is removed from an archetype table, its slot is filled by
// moving the last entity in the same chunk into the freed position. This
// swap-and-pop strategy avoids holes in memory, maintains dense storage,
// and eliminates background defragmentation. The caller is responsible for
// updating the entity's location in the entity index after a swap.
//
// # Archetype Creation Callback
//
// [Catalog.Init] accepts a func(*Archetype) callback invoked each time a new
// archetype is registered. The caller wires in whatever notification logic it
// needs without creating a circular import. This call is on the cold path and
// carries no performance cost during normal iteration.
package arch
