package query

import (
	"iter"
	"unsafe"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/uid"
)

// view0 uses a phantom type parameter so the compiler performs escape analysis
// at each call site (as with generic View1-10), enabling zero-alloc iteration.
type view0[_ any] struct {
	*View
}

// View0 iterates entities matching a filter without reading any component data.
type View0 = view0[struct{}]

// NewView0 creates a View for entities filtered by opts, with no component columns.
func NewView0(catalog *Catalog, opts ...comp.BlueprintOpt) *View0 {
	var s comp.Spec0
	s.Init(catalog.cc, opts...)
	return &view0[struct{}]{View: catalog.AddView(&s.Blueprint)}
}

// All returns an iterator over matched entities, yielded in contiguous chunks.
func (v *view0[_]) All() iter.Seq[struct{ Entity []uid.UID64 }] {
	return func(yield func(struct{ Entity []uid.UID64 }) bool) {
		for _, bt := range v.BakedTables {
			for i := range bt.Table.Chunks {
				chunk := &bt.Table.Chunks[i]
				count := chunk.Len
				if count == 0 {
					continue
				}
				base := chunk.Ptr
				if !yield(struct{ Entity []uid.UID64 }{
					Entity: unsafe.Slice((*uid.UID64)(base), count),
				}) {
					return
				}
			}
		}
	}
}

// Filter yields the subset of selected entities that match the View's filter.
func (v *view0[_]) Filter(selected []uid.UID64) iter.Seq2[int, struct{ Entity uid.UID64 }] {
	return func(yield func(int, struct{ Entity uid.UID64 }) bool) {
		entityIndex := v.EntityIndex
		var lastArchID arch.ID = arch.NullID
		var bt *BakedTable
		for i, e := range selected {
			link, ok := entityIndex.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				bt = v.Get(link.ArchId)
				lastArchID = link.ArchId
			}
			if bt == nil {
				continue
			}
			if !yield(i, struct{ Entity uid.UID64 }{Entity: e}) {
				return
			}
		}
	}
}
