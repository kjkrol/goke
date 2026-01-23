package ecs

type ViewBuilder struct {
	reg     *Registry
	compIDs []ComponentID
}

func NewViewBuilder(reg *Registry) *ViewBuilder {
	return &ViewBuilder{
		reg:     reg,
		compIDs: make([]ComponentID, 0, MaxComponents),
	}
}

func OnCompType[T any](b *ViewBuilder) {
	id := ensureComponentRegistered[T](b.reg.componentsRegistry)
	b.OnType(id)
}

func (b *ViewBuilder) OnType(id ComponentID) *ViewBuilder {
	b.compIDs = append(b.compIDs, id)
	return b
}

func (b *ViewBuilder) Build() *View {
	var mask ArchetypeMask
	for _, id := range b.compIDs {
		mask = mask.Set(id)
	}

	v := &View{
		reg:     b.reg,
		mask:    mask,
		compIDs: b.compIDs,
	}
	v.Reindex()
	return v
}
