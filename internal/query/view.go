package query

import (
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
)

type View struct {
	BakedTablesCatalog
	EntityIndex *ent.Index
	composition comp.Composition
	excludeMask comp.Mask
}

func (v *View) Init(entityIndex *ent.Index, blueprint *comp.Blueprint) {
	includeMask := comp.Mask{}.Build(blueprint)

	var excludeMask comp.Mask
	for _, id := range blueprint.ExCompIDs {
		if includeMask.IsSet(id) {
			panic("ECS View Error: component cannot be both REQUIRED and EXCLUDED")
		}
		excludeMask = excludeMask.Set(id)
	}

	v.EntityIndex = entityIndex
	v.composition = comp.Composition{Mask: includeMask, Metas: blueprint.CompInfos}
	v.excludeMask = excludeMask
}

func (v *View) Clear() {
	v.EntityIndex = nil
	v.composition = comp.Composition{}
	v.excludeMask = comp.Mask{}
	v.BakedTablesCatalog.Clear()
}

func (v *View) BakeIfMatch(archetype *arch.Archetype) {
	if archetype.Table.NumColumns() > 0 && v.Matches(archetype.Mask()) {
		v.Bake(archetype)
	}
}

func (v *View) Bake(archetype *arch.Archetype) {
	if archetype.Table.NumColumns() == 0 {
		return
	}
	v.BakedTablesCatalog.Add(archetype, v.composition.Metas)
}

func (v *View) Matches(archMask comp.Mask) bool {
	return archMask.Matches(v.composition.Mask, v.excludeMask)
}
