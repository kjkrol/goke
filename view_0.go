package goke

import (
	"iter"

	"github.com/kjkrol/goke/internal/core"
)

type View0 struct {
	*core.View
}

func NewView0(ecs *ECS, opts ...BlueprintOption) *View0 {
	blueprint := core.NewBlueprint(ecs.registry)
	for _, opt := range opts {
		opt(blueprint)
	}
	view := core.NewView(blueprint, ecs.registry)
	return &View0{View: view}
}

func (v *View0) All() iter.Seq[core.Entity] {
	return func(yield func(core.Entity) bool) {
		for i := range v.Baked {
			b := &v.Baked[i]
			for j := 0; j < b.GetLen(); j++ {
				if !yield((b.GetEntities())[j]) {
					return
				}
			}
		}
	}
}

func (v *View0) Filter(selected []Entity) iter.Seq[core.Entity] {
	links := v.Reg.ArchetypeRegistry.EntityLinkStore
	return func(yield func(core.Entity) bool) {
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
