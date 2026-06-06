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

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//
//	for page := range view0.All() {
//		for i, entity := range page.Entity {
//			// Apply domain logic here...
//		}
//	}
func (v *View0) All() iter.Seq[struct{ Entity []Entity }] {
	return func(yield func(struct{ Entity []Entity }) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
					}) {
					return
				}
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
	return func(yield func(struct{ Entity core.Entity }, struct{}) bool) {
		var lastArchID int32 = -1
		registry := v.Reg.ArchetypeRegistry
		for _, e := range selected {
			link, ok := registry.EntityLinkStore.Get(e)
			if !ok {
				continue
			}

			if int32(link.ArchId) != lastArchID {
				arch := &registry.Archetypes[link.ArchId]
				if !v.View.Matches(arch.Mask) {
					lastArchID = -1
					continue
				}

				lastArchID = int32(link.ArchId)
			}

			head := struct{ Entity core.Entity }{Entity: e}

			if !yield(head, struct{}{}) {
				return
			}
		}
	}
}
