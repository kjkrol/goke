// Package reg is the top-level ECS facade.
//
// [Registry] wires three subsystems into a single world:
//
//   - [ent.Manager]    — entity lifecycle: address book ([addr.Book]) and
//     archetype graph management
//   - [comp.MetaIndex] — maps Go types to stable [comp.Meta] descriptors
//   - [query.Catalog]  — maintains active views as new archetypes are created
//
// All mutation methods on [Registry] are thin delegates to the subsystem that
// owns the operation. [Registry.AddView] is the one coordinating method: it
// registers a new view with the query catalog and immediately bakes it against
// all archetypes already in the archetype catalog.
package reg
