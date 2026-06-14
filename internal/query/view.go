package query

import (
	"fmt"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
)

const EntitySize = unsafe.Sizeof(uid.UID64(0))

type View struct {
	ArchReg      *arch.Catalog
	includeMask  comp.Mask
	excludeMask  comp.Mask
	Layout       []comp.Meta
	MatchedArchs []MatchedArch
	archMapping  []int32
}

func (v *View) Clear() {
	v.includeMask = comp.Mask{}
	v.excludeMask = comp.Mask{}
	clear(v.Layout)
	v.Layout = nil
	clear(v.MatchedArchs)
	v.MatchedArchs = nil
	clear(v.archMapping)
	v.archMapping = nil
}

func NewView(
	blueprint *comp.Blueprint,
	layout []comp.Meta,
	archReg *arch.Catalog,
	viewReg *Registry,
) *View {
	includeMask := comp.Mask{}.Build(blueprint)

	var excludeMask comp.Mask
	for _, id := range blueprint.ExCompIDs {
		if includeMask.IsSet(id) {
			panic("ECS View Error: component cannot be both REQUIRED and EXCLUDED")
		}
		excludeMask = excludeMask.Set(id)
	}

	for _, info := range layout {
		if !includeMask.IsSet(info.ID) {
			panic(fmt.Sprintf("View Layout Error: Component %d is in layout but not required by Blueprint", info.ID))
		}
	}

	v := &View{
		ArchReg:     archReg,
		includeMask: includeMask,
		excludeMask: excludeMask,
		Layout:      layout,
	}
	v.Reindex()
	viewReg.Register(v)
	return v
}

func (v *View) Reindex() {
	v.MatchedArchs = v.MatchedArchs[:0]
	r := v.ArchReg

	requiredLen := int(r.Len())
	if cap(v.archMapping) < requiredLen {
		v.archMapping = make([]int32, requiredLen)
	}
	v.archMapping = v.archMapping[:requiredLen]
	for i := range v.archMapping {
		v.archMapping[i] = -1
	}

	for i := arch.RootID; i < r.Len(); i++ {
		v.AddArchetypeIfMatch(&r.Archetypes[i])
	}
}

func (v *View) AddArchetypeIfMatch(a *arch.Archetype) {
	if a.Table.NumColumns() > 0 && v.Matches(a.Mask()) {
		v.AddArchetype(a)
	}
}

func (v *View) AddArchetype(a *arch.Archetype) {
	if a.Table.NumColumns() == 0 {
		return
	}

	offsets := make([]uintptr, len(v.Layout))
	sizes := make([]uintptr, len(v.Layout))
	for i, info := range v.Layout {
		col := a.Table.GetColumn(info.ID)
		if col == nil {
			continue
		}
		offsets[i] = col.PageOffset
		sizes[i] = col.ItemSize
	}

	v.MatchedArchs = append(v.MatchedArchs, MatchedArch{
		Table:            &a.Table,
		EntityPageOffset: a.Table.GetEntityColumn().PageOffset,
		CompOffsets:      offsets,
		CompSizes:        sizes,
	})

	if int(a.Id) >= len(v.archMapping) {
		v.growArchMapping(int(a.Id) + 1)
	}
	v.archMapping[a.Id] = int32(len(v.MatchedArchs) - 1)
}

func (v *View) growArchMapping(minLen int) {
	newCap := max(cap(v.archMapping)*2, minLen)
	grown := make([]int32, minLen, newCap)
	copy(grown, v.archMapping)
	for i := len(v.archMapping); i < minLen; i++ {
		grown[i] = -1
	}
	v.archMapping = grown
}

func (v *View) Matches(archMask comp.Mask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}

func (v *View) GetMatchedArch(id arch.ID) *MatchedArch {
	if int(id) >= len(v.archMapping) {
		return nil
	}

	idx := v.archMapping[id]
	if idx == -1 {
		return nil
	}

	return &v.MatchedArchs[idx]
}
