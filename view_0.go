package goke

import (
	"iter"
	"unsafe"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/uid"
)

// View0 provides a high-performance iterator for entities that match
// a specific architectural mask but does not require access to any
// specific component data columns.
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
			for i := range ma.Arch.Memory.Pages {
				page := &ma.Arch.Memory.Pages[i]
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

// Filter iterates `selected` entities and yields one entity at a time
// for those that match the View's archetype constraints.
func (v *View0) Filter(selected []uid.UID64) iter.Seq2[int, struct{ Entity uid.UID64 }] {
	return func(yield func(int, struct{ Entity uid.UID64 }) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			if !yield(i, struct{ Entity uid.UID64 }{Entity: e}) {
				return
			}
		}
	}
}
