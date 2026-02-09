package goke

import (
	"iter"
	"unsafe"

	"github.com/kjkrol/goke/internal/core"
)

// View0 provides a high-performance iterator for entities that match
// a specific architectural mask but does not require access to any
// specific component data columns.
//
// It is primarily used for:
// 1. Tag-only queries (finding entities that have a specific Tag).
// 2. Systems that only need the Entity ID.
// 3. Counting entities matching a criteria.
type View0 struct {
	*core.View
}

// NewView0 initializes a query for entities matching the provided options,
// without fetching any component data.
//
// Example:
//
//	// Find all entities with "EnemyTag"
//	view := goke.NewView0(ecs, goke.WithTag(EnemyTag))
func NewView0(ecs *ECS, opts ...BlueprintOption) *View0 {
	blueprint := core.NewBlueprint(ecs.registry)
	for _, opt := range opts {
		opt(blueprint)
	}
	// Empty component slice because View0 doesn't read component data
	view := core.NewView(blueprint, []core.ComponentInfo{}, ecs.registry)
	return &View0{View: view}
}

// All returns an iterator (iter.Seq2) that yields the unique Entity identifier.
// Even though View0 does not access component data, it iterates over
// archetypes linearly, ensuring maximum cache efficiency when reading Entity IDs.
//
// Example usage:
//
//	for head, _ := range view0.All() {
//	    entity := head.Entity
//	    // Logic using only the entity ID...
//	}
func (v *View0) All() iter.Seq2[struct{ Entity core.Entity }, struct{}] {
	return func(yield func(struct{ Entity core.Entity }, struct{}) bool) {
		stride := unsafe.Sizeof(core.Entity(0))

		for i := range v.Baked {
			b := &v.Baked[i]

			count := *b.Len
			if count == 0 {
				continue
			}

			// FIX: Access EntityColumn directly.
			// b.Columns is empty for View0, so b.Columns[0] would panic.
			ptr := b.EntityColumn.Data

			for j := 0; j < count; j++ {
				val := *(*core.Entity)(ptr)

				head := struct{ Entity core.Entity }{Entity: val}

				if !yield(head, struct{}{}) {
					return
				}

				ptr = unsafe.Add(ptr, stride)
			}
		}
	}
}

// Filter returns an iterator (iter.Seq2) that yields only the entities
// specified in the selected slice, provided they match the View's archetype
// constraints.
//
// The iterator performs an internal validation for each entity to ensure
// it still belongs to an archetype compatible with this View.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for head, _ := range view0.Filter(selected) {
//	    entity := head.Entity
//	    // ...
//	}
func (v *View0) Filter(selected []core.Entity) iter.Seq2[struct{ Entity core.Entity }, struct{}] {
	links := v.Reg.ArchetypeRegistry.EntityLinkStore

	return func(yield func(struct{ Entity core.Entity }, struct{}) bool) {
		for _, e := range selected {
			link, ok := links.Get(e)
			if !ok {
				continue
			}
			// Unsafe access to Archetypes array for speed (bounds checked implicitly by LinkStore logic)
			arch := &v.Reg.ArchetypeRegistry.Archetypes[link.ArchId]

			// Check if the entity's current archetype matches this View
			if !v.Matches(arch.Mask) {
				continue
			}

			head := struct{ Entity core.Entity }{Entity: e}

			if !yield(head, struct{}{}) {
				return
			}
		}
	}
}
