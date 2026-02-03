package goke

import (
	"iter"

	core2 "github.com/kjkrol/goke/internal/core"
)

type View0 struct {
	*core2.View
}

func NewView0(eng *Engine, options ...core2.ViewOption) *View0 {
	view := core2.NewView(eng.registry, options...)
	return &View0{View: view}
}

func (v *View0) All() iter.Seq[core2.Entity] {
	return func(yield func(core2.Entity) bool) {
		for i := range v.Baked {
			b := &v.Baked[i]
			for j := 0; j < *b.Len; j++ {
				if !yield((*b.Entities)[j]) {
					return
				}
			}
		}
	}
}

func (v *View0) Filter(selected []Entity) iter.Seq[core2.Entity] {
	links := v.Reg.ArchetypeRegistry.EntityLinkStore
	return func(yield func(core2.Entity) bool) {
		for _, e := range selected {
			link, ok := links.Get(e)
			if !ok {
				continue
			}
			arch := link.Arch

			if arch == nil || !v.Matches(arch.Mask) {
				continue
			}

			if !yield(e) {
				return
			}
		}
	}
}
