package query

import "github.com/kjkrol/goke/internal/arch"

type Registry struct {
	views []*View
}

func NewRegistry(initialCap int) Registry {
	return Registry{views: make([]*View, 0, initialCap)}
}

func (vr *Registry) Register(v *View) {
	vr.views = append(vr.views, v)
}

func (vr *Registry) OnArchetypeCreated(a *arch.Archetype) {
	for _, v := range vr.views {
		if v.Matches(a.Mask()) {
			v.AddArchetype(a)
		}
	}
}

func (vr *Registry) Reset() {
	for i := range vr.views {
		if vr.views[i] != nil {
			vr.views[i].Clear()
		}
	}
	clear(vr.views)
	vr.views = vr.views[:0]
}

var _ arch.Observer = (*Registry)(nil)
