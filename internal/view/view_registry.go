package view

import "github.com/kjkrol/goke/internal/arch"

type ViewRegistry struct {
	views []*View
}

func NewViewRegistry(initialCap int) ViewRegistry {
	return ViewRegistry{views: make([]*View, 0, initialCap)}
}

func (vr *ViewRegistry) Register(v *View) {
	vr.views = append(vr.views, v)
}

func (vr *ViewRegistry) OnArchetypeCreated(a *arch.Archetype) {
	for _, v := range vr.views {
		if v.Matches(a.Mask) {
			v.AddArchetype(a)
		}
	}
}

func (vr *ViewRegistry) Reset() {
	for i := range vr.views {
		if vr.views[i] != nil {
			vr.views[i].Clear()
		}
	}
	clear(vr.views)
	vr.views = vr.views[:0]
}

var _ arch.ArchetypeObserver = (*ViewRegistry)(nil)
