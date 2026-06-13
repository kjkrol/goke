// Package reg is the top-level ECS facade.
//
// [Registry] wires three subsystems into a single world:
//
//   - [entity.Manager] — owns entity lifecycle, component composition,
//     and the underlying archetype graph (see [entity] for details)
//   - [comp.Catalog]   — maps Go types to stable [comp.Meta] descriptors
//   - [query.Catalog]  — maintains active views as new archetypes are created
//
// All mutation methods on [Registry] are thin delegates to the subsystem that
// owns the operation. [Registry.AddView] is the one coordinating method: it
// registers a new view with the query catalog and immediately bakes it against
// all archetypes already known to the entity manager.
package reg
