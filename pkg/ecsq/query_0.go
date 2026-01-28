package ecsq

import (
	"iter"

	"github.com/kjkrol/goke/internal/core"
)

type Query0 struct {
	*core.View
}

func NewQuery0(reg *core.Registry, options ...core.ViewOption) *Query0 {
	view := core.NewView(reg, options...)
	return &Query0{View: view}
}

func (q *Query0) All() iter.Seq[core.Entity] {
	return func(yield func(core.Entity) bool) {
		for i := range q.Baked {
			b := &q.Baked[i]
			for j := 0; j < *b.Len; j++ {
				if !yield((*b.Entities)[j]) {
					return
				}
			}
		}
	}
}

func (q *Query0) Filter(entities []core.Entity) iter.Seq[core.Entity] {
	links := q.Reg.ArchetypeRegistry.EntityArchLinks
	return func(yield func(core.Entity) bool) {
		var lastArch *core.Archetype
		var cols [0]*core.Column
		for _, e := range entities {
			link := links[e.Index()]
			arch := link.Arch
			if arch == nil || !q.View.Matches(arch.Mask) {
				continue
			}
			if arch != lastArch {
				for i := 0; i < 0; i++ {
					cols[i] = arch.Columns[q.CompIDs[i]]
				}
				lastArch = arch
			}
			if !yield(e) {
				return
			}
		}
	}
}
