package ecs

type ViewRegistry struct {
	views []*View
}

func NewViewRegistry(initCapacity int) *ViewRegistry {
	return &ViewRegistry{views: make([]*View, 0, initCapacity)}
}

func (vr *ViewRegistry) Register(v *View) {
	vr.views = append(vr.views, v)
}

func (vr *ViewRegistry) OnArchetypeCreated(arch *Archetype) {
	for _, v := range vr.views {
		if arch.mask.Contains(v.mask) {
			v.AddArchetype(arch)
		}
	}
}
