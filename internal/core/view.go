package core

import (
	"fmt"
	"unsafe"
)

const EntitySize = unsafe.Sizeof(Entity(0))

type MatchedArch struct {
	Arch              *Archetype
	EntityChunkOffset uintptr
	FieldsOffsets     []uintptr
	FieldsSizes       []uintptr
}

func (ma *MatchedArch) Clear() {
	ma.Arch = nil
	clear(ma.FieldsOffsets)
	ma.FieldsOffsets = nil
	clear(ma.FieldsSizes)
	ma.FieldsSizes = nil
	ma.EntityChunkOffset = 0
}

type View struct {
	Reg         *Registry
	includeMask ArchetypeMask
	excludeMask ArchetypeMask
	Layout      []ComponentInfo
	Baked       []MatchedArch
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
		oldArchStruct := &v.Baked[len(v.Baked)]

		// Check if the recycled slices are big enough for current layout
		if cap(oldArchStruct.FieldsOffsets) >= len(v.Layout) {
			offsets = oldArchStruct.FieldsOffsets[:len(v.Layout)]
			sizes = oldArchStruct.FieldsSizes[:len(v.Layout)]
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
	// We copy ChunkOffset and ItemSize from the Archetype into our local arrays.
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
		offsets[i] = col.ChunkOffset
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
		Arch:              arch,
		EntityChunkOffset: entCol.ChunkOffset,
		FieldsOffsets:     offsets,
		FieldsSizes:       sizes,
	})
}

func (v *View) Matches(archMask ArchetypeMask) bool {
	return archMask.Matches(v.includeMask, v.excludeMask)
}
