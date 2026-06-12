package view

import (
	"fmt"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/internal/mem"
)

const EntitySize = unsafe.Sizeof(uid.UID64(0))

type View struct {
	ArchReg     *arch.ArchetypeRegistry
	includeMask core.ArchetypeMask
	excludeMask core.ArchetypeMask
	Layout      []core.ComponentInfo
	Baked       []MatchedArch
	archMapping []int32
}

func (v *View) Clear() {
	v.includeMask = core.ArchetypeMask{}
	v.excludeMask = core.ArchetypeMask{}
	clear(v.Layout)
	v.Layout = nil
	for i := range v.Baked {
		v.Baked[i].Clear()
	}
	clear(v.Baked)
	v.Baked = nil
	clear(v.archMapping)
	v.archMapping = nil
}

func NewView(
	includeMask core.ArchetypeMask,
	excludeMask core.ArchetypeMask,
	layout []core.ComponentInfo,
	archReg *arch.ArchetypeRegistry,
	viewReg *ViewRegistry,
) *View {
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
	v.Baked = v.Baked[:0]
	r := v.ArchReg

	requiredLen := int(r.LastArchetypeId())
	if cap(v.archMapping) < requiredLen {
		v.archMapping = make([]int32, requiredLen)
	}
	v.archMapping = v.archMapping[:requiredLen]
	for i := range v.archMapping {
		v.archMapping[i] = -1
	}

	for i := core.RootArchetypeId; i < r.LastArchetypeId(); i++ {
		v.AddArchetypeIfMatch(&r.Archetypes[i])
	}
}

func (v *View) AddArchetypeIfMatch(a *arch.Archetype) {
	if len(a.Columns) > 0 && v.Matches(a.Mask) {
		v.AddArchetype(a)
	}
}

func (v *View) AddArchetype(a *arch.Archetype) {
	if len(a.Columns) == 0 {
		return
	}

	var offsets []uintptr
	var sizes []uintptr

	if cap(v.Baked) > len(v.Baked) {
		extended := v.Baked[:len(v.Baked)+1]
		oldArchStruct := &extended[len(v.Baked)]

		if cap(oldArchStruct.CompOffsets) >= len(v.Layout) {
			offsets = oldArchStruct.CompOffsets[:len(v.Layout)]
			sizes = oldArchStruct.CompSizes[:len(v.Layout)]
		}
	}

	if offsets == nil && len(v.Layout) > 0 {
		offsets = make([]uintptr, len(v.Layout))
		sizes = make([]uintptr, len(v.Layout))
	}

	for i, info := range v.Layout {
		localIdx := a.Map[info.ID]

		if localIdx == mem.InvalidLocalID || int(localIdx) >= len(a.Columns) {
			continue
		}

		col := &a.Columns[localIdx]
		offsets[i] = col.PageOffset
		sizes[i] = col.ItemSize
	}

	entCol := a.GetEntityColumn()

	v.Baked = append(v.Baked, MatchedArch{
		Arch:             a,
		EntityPageOffset: entCol.PageOffset,
		CompOffsets:      offsets,
		CompSizes:        sizes,
	})

	if int(a.Id) >= len(v.archMapping) {
		v.growArchMapping(int(a.Id) + 1)
	}
	v.archMapping[a.Id] = int32(len(v.Baked) - 1)
}

func (v *View) growArchMapping(minLen int) {
	newCap := cap(v.archMapping) * 2
	if newCap < minLen {
		newCap = minLen
	}
	grown := make([]int32, minLen, newCap)
	copy(grown, v.archMapping)
	for i := len(v.archMapping); i < minLen; i++ {
		grown[i] = -1
	}
	v.archMapping = grown
}

func (v *View) Matches(archMask core.ArchetypeMask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}

func (v *View) GetMatchedArch(id core.ArchetypeId) *MatchedArch {
	if int(id) >= len(v.archMapping) {
		return nil
	}

	idx := v.archMapping[id]
	if idx == -1 {
		return nil
	}

	return &v.Baked[idx]
}

