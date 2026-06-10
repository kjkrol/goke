package core

import "github.com/kjkrol/uid"

import (
	"fmt"
	"unsafe"
)

const EntitySize = unsafe.Sizeof(uid.UID64(0))

type MatchedArch struct {
	Arch             *Archetype
	EntityPageOffset uintptr
	CompOffsets      []uintptr
	CompSizes        []uintptr
}

func (ma *MatchedArch) Clear() {
	ma.Arch = nil
	clear(ma.CompOffsets)
	ma.CompOffsets = nil
	clear(ma.CompSizes)
	ma.CompSizes = nil
	ma.EntityPageOffset = 0
}

type View struct {
	Reg         *Registry
	includeMask ArchetypeMask
	excludeMask ArchetypeMask
	Layout      []ComponentInfo
	Baked       []MatchedArch
	archMapping []int32
}

func (v *View) Clear() {
	v.includeMask = ArchetypeMask{}
	v.excludeMask = ArchetypeMask{}
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

// View factory based on Functional Options pattern
func NewView(blueprint *Blueprint, layout []ComponentInfo, reg *Registry) *View {
	var mask ArchetypeMask
	var excludedMask ArchetypeMask

	for _, info := range blueprint.compInfos {
		mask = mask.Set(info.ID)
	}

	for _, id := range blueprint.tagIDs {
		mask = mask.Set(id)
	}

	for _, id := range blueprint.exCompIDs {
		if mask.IsSet(id) {
			panic(fmt.Sprintf("ECS View Error: Component ID %d cannot be both REQUIRED and EXCLUDED", id))
		}
		excludedMask = excludedMask.Set(id)
	}

	for _, info := range layout {
		if !mask.IsSet(info.ID) {
			panic(fmt.Sprintf("View Layout Error: Component %d is in layout but not required by Blueprint", info.ID))
		}
	}

	v := &View{
		Reg:         reg,
		includeMask: mask,
		excludeMask: excludedMask,
		Layout:      layout,
	}
	v.Reindex()
	v.Reg.ViewRegistry.Register(v)
	return v
}

func (v *View) Reindex() {
	v.Baked = v.Baked[:0]
	reg := v.Reg.ArchetypeRegistry

	requiredLen := int(reg.lastArchetypeId)
	if cap(v.archMapping) < requiredLen {
		v.archMapping = make([]int32, requiredLen)
	}
	v.archMapping = v.archMapping[:requiredLen]
	for i := range v.archMapping {
		v.archMapping[i] = -1
	}

	for i := RootArchetypeId; i < reg.lastArchetypeId; i++ {
		v.AddArchetypeIfMatch(&reg.Archetypes[i])
	}
}

func (v *View) AddArchetypeIfMatch(arch *Archetype) {
	if len(arch.Columns) > 0 && v.Matches(arch.Mask) {
		v.AddArchetype(arch)
	}
}

func (v *View) AddArchetype(arch *Archetype) {
	// 1. Safety Check: Skip empty archetypes
	if len(arch.Columns) == 0 {
		return
	}

	// -------------------------------------------------------------------------
	// STEP 1: Memory Reuse Strategy (Zero Allocation Trick)
	// -------------------------------------------------------------------------
	// If the View was cleared using `v.Baked = v.Baked[:0]`, the underlying array
	// still exists and holds old MatchedArch structs. We can steal their
	// allocated slices (FieldsOffsets/FieldsSizes) to avoid new allocations.

	var offsets []uintptr
	var sizes []uintptr

	// Check if there is "hidden" capacity in the Baked slice
	if cap(v.Baked) > len(v.Baked) {
		// Access the "garbage" element that is about to be overwritten
		extended := v.Baked[:len(v.Baked)+1]
		oldArchStruct := &extended[len(v.Baked)]

		// Check if the recycled slices are big enough for current layout
		if cap(oldArchStruct.CompOffsets) >= len(v.Layout) {
			offsets = oldArchStruct.CompOffsets[:len(v.Layout)]
			sizes = oldArchStruct.CompSizes[:len(v.Layout)]
		}
	}

	// If we couldn't reuse memory (first run or layout changed), allocate new.
	if offsets == nil && len(v.Layout) > 0 {
		offsets = make([]uintptr, len(v.Layout))
		sizes = make([]uintptr, len(v.Layout))
	}

	// -------------------------------------------------------------------------
	// STEP 2: Value Caching (Flattening the Data)
	// -------------------------------------------------------------------------
	// We copy PageOffset and ItemSize from the Archetype into our local arrays.
	// This allows the View Iterator to calculate pointers purely using math,
	// without ever dereferencing the `*Column` pointer or touching `Arch` memory.

	for i, info := range v.Layout {
		// Map lookup: Global Component ID -> Local Archetype Index
		localIdx := arch.Map[info.ID]

		// Safety check (should be covered by Mask matching, but safety first)
		if localIdx == InvalidLocalID || int(localIdx) >= len(arch.Columns) {
			continue
		}

		// Direct pointer to the immutable column definition in Archetype
		col := &arch.Columns[localIdx]

		// COPY VALUES to local cache
		offsets[i] = col.PageOffset
		sizes[i] = col.ItemSize
	}

	// -------------------------------------------------------------------------
	// STEP 3: Entity Cache & Finalize
	// -------------------------------------------------------------------------
	// Grab the Entity Column Offset for the fastest possible Entity iteration.
	entCol := arch.GetEntityColumn()

	// Append the fully optimized structure.
	// Note: This might overwrite an old struct in the underlying array,
	// effectively "recycling" the slice headers we just prepared.
	v.Baked = append(v.Baked, MatchedArch{
		Arch:             arch,
		EntityPageOffset: entCol.PageOffset,
		CompOffsets:      offsets,
		CompSizes:        sizes,
	})

	// Ensure archMapping can hold the new archetype's id. This grows the
	// mapping when the view receives a freshly created archetype via
	// ViewRegistry.OnArchetypeCreated after the initial Reindex sized the
	// slice to the then-current lastArchetypeId.
	if int(arch.Id) >= len(v.archMapping) {
		v.growArchMapping(int(arch.Id) + 1)
	}
	v.archMapping[arch.Id] = int32(len(v.Baked) - 1)
}

// growArchMapping extends archMapping to at least `minLen` entries, padding
// new slots with -1 ("no matched arch yet"). Capacity doubling keeps amortized
// cost constant when many archetypes are appended in sequence.
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

func (v *View) Matches(archMask ArchetypeMask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}

func (v *View) GetMatchedArch(id ArchetypeId) *MatchedArch {
	if int(id) >= len(v.archMapping) {
		return nil
	}

	idx := v.archMapping[id]
	if idx == -1 {
		return nil
	}

	return &v.Baked[idx]
}
