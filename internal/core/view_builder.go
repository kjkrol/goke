package core

import "fmt"

type ViewBuilder struct {
	*BlueprintBuilder
	excludedCompIDs []ComponentID
}

func NewViewBuilder(reg *Registry) *ViewBuilder {
	return &ViewBuilder{
		BlueprintBuilder: NewBlueprintBuilder(reg),
		excludedCompIDs:  make([]ComponentID, 0, MaxColumns),
	}
}

func OnCompType[T any](b *ViewBuilder) {
	compInfo := EnsureComponentRegistered[T](b.reg.ComponentsRegistry)
	b.WithComp(compInfo.ID)
}

func OnTagType[T any](b *ViewBuilder) {
	compInfo := EnsureComponentRegistered[T](b.reg.ComponentsRegistry)
	b.WithTag(compInfo.ID)
}

func OnCompExcludeType[T any](b *ViewBuilder) {
	compInfo := EnsureComponentRegistered[T](b.reg.ComponentsRegistry)
	b.OnExcludeType(compInfo.ID)
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
		Reg:         b.reg,
		includeMask: mask,
		excludeMask: excludedMask,
		CompIDs:     b.compIDs,
	}
	v.Reindex()
	v.Reg.ViewRegistry.Register(v)
	return v
}
