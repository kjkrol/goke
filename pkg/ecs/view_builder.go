package ecs

import "fmt"

type ViewBuilder struct {
	reg             *Registry
	compIDs         []ComponentID
	tagIDs          []ComponentID
	excludedCompIDs []ComponentID
}

func NewViewBuilder(reg *Registry) *ViewBuilder {
	return &ViewBuilder{
		reg:             reg,
		compIDs:         make([]ComponentID, 0, MaxComponents),
		tagIDs:          make([]ComponentID, 0, MaxComponents),
		excludedCompIDs: make([]ComponentID, 0, MaxComponents),
	}
}

func OnCompType[T any](b *ViewBuilder) {
	compInfo := ensureComponentRegistered[T](b.reg.componentsRegistry)
	b.OnType(compInfo.ID)
}

func OnTagType[T any](b *ViewBuilder) {
	compInfo := ensureComponentRegistered[T](b.reg.componentsRegistry)
	b.OnTag(compInfo.ID)
}

func OnCompExcludeType[T any](b *ViewBuilder) {
	compInfo := ensureComponentRegistered[T](b.reg.componentsRegistry)
	b.OnExcludeType(compInfo.ID)
}

func (b *ViewBuilder) OnType(id ComponentID) *ViewBuilder {
	b.compIDs = append(b.compIDs, id)
	return b
}

func (b *ViewBuilder) OnTag(id ComponentID) *ViewBuilder {
	b.tagIDs = append(b.tagIDs, id)
	return b
}

func (b *ViewBuilder) OnExcludeType(id ComponentID) *ViewBuilder {
	b.excludedCompIDs = append(b.excludedCompIDs, id)
	return b
}

func (b *ViewBuilder) Build() *View {
	var mask ArchetypeMask
	var excludedMask ArchetypeMask

	for _, id := range b.compIDs {
		mask = mask.Set(id)
	}

	for _, id := range b.tagIDs {
		mask = mask.Set(id)
	}

	for _, id := range b.excludedCompIDs {
		if mask.IsSet(id) {
			panic(fmt.Sprintf("ECS View Error: Component ID %d cannot be both REQUIRED and EXCLUDED", id))
		}
		excludedMask = excludedMask.Set(id)
	}

	v := &View{
		reg:         b.reg,
		includeMask: mask,
		excludeMask: excludedMask,
		compIDs:     b.compIDs,
	}
	v.Reindex()
	v.reg.viewRegistry.Register(v)
	return v
}
