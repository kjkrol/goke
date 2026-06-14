package goke

import (
	"iter"
	"unsafe"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/query"
)

// View0 iterates entities matching a filter without reading any component data.
// Useful for tag-only queries (e.g. "all enemies").
type View0 struct {
	*query.View
}

// NewView0 creates a View that matches entities according to the provided BlueprintOptions,
// without fetching any component columns.
//
// Example:
//
//	view := goke.NewView0(ecs, goke.Include[EnemyTag]())
func NewView0(ecs *ECS, opts ...BlueprintOption) *View0 {
	blueprint := comp.NewBlueprint()
	for _, opt := range opts {
		opt(blueprint, &ecs.registry)
	}
	// Empty component slice because View0 doesn't read component data
	v := query.NewView(blueprint, []comp.Meta{}, &ecs.registry.ArchCatalog, &ecs.registry.ViewRegistry)
	return &View0{View: v}
}

// All returns an iterator over matched entities, yielded in contiguous chunks.
//
// Example:
//
//	for chunk := range view0.All() {
//		for _, entity := range chunk.Entity {
//			// process entity
//		}
//	}
func (v *View0) All() iter.Seq[struct{ Entity []EntityID }] {
	return func(yield func(struct{ Entity []EntityID }) bool) {
		for _, ma := range v.MatchedArchs {
			for i := range ma.Table.Chunks {
				chunk := &ma.Table.Chunks[i]
				count := chunk.Len
				if count == 0 {
					continue
				}
				base := chunk.Ptr
				if !yield(
					struct {
						Entity []EntityID
					}{
						Entity: unsafe.Slice((*EntityID)(unsafe.Add(base, ma.EntityPageOffset)), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter yields the subset of selected entities that match the View's filter.
func (v *View0) Filter(selected []EntityID) iter.Seq2[int, struct{ Entity EntityID }] {
	return func(yield func(int, struct{ Entity EntityID }) bool) {
		store := &v.ArchReg.EntityIndex

		var lastArchID arch.ID = arch.NullID
		var ma *query.MatchedArch
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
			if !yield(i, struct{ Entity EntityID }{Entity: e}) {
				return
			}
		}
	}
}
