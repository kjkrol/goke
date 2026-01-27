package ecs

type ViewRegistry struct {
	views []*View
}

func NewViewRegistry() *ViewRegistry {
	return &ViewRegistry{views: make([]*View, 0, viewRegistryInitCap)}
}

func (vr *ViewRegistry) Register(v *View) {
	vr.views = append(vr.views, v)
}

func (vr *ViewRegistry) OnArchetypeCreated(arch *Archetype) {
	for _, v := range vr.views {
		if v.matches(arch.mask) {
			v.AddArchetype(arch)
		}
	}
}
