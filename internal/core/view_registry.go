package core

type ViewRegistry struct {
	views []*View
}

func NewViewRegistry(viewRegistryInitCap int) ViewRegistry {
	return ViewRegistry{views: make([]*View, 0, viewRegistryInitCap)}
}

func (vr *ViewRegistry) Register(v *View) {
	vr.views = append(vr.views, v)
}

func (vr *ViewRegistry) OnArchetypeCreated(arch *Archetype) {
	for _, view := range vr.views {
		if view.Matches(arch.Mask) {
			view.AddArchetype(arch)
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
