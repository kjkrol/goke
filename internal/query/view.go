package query

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/addr"
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

type iterMode uint8

const (
	modeNone iterMode = iota
	modeAll
	modeFilter
)

type allIter struct {
	tableIdx int
	chunkIdx int
}

type filterIter struct {
	selected   []uid.UID64
	pos        int
	lastArchID arch.ID
	bt         *BakedTable
	Entity     uid.UID64
	Idx        int
}

type View struct {
	BakedTablesCatalog
	EntityIndex *addr.Index
	includeMask comp.Mask
	compIDs     []comp.ID
	excludeMask comp.Mask
	mode        iterMode
	Cursor      iter.Cursor
	allIter
	filterIter
}

func (v *View) Init(entityIndex *addr.Index, blueprint *comp.Blueprint) {
	includeMask := comp.NewMask(blueprint)

	var excludeMask comp.Mask
	for _, id := range blueprint.ExCompIDs {
		if includeMask.IsSet(id) {
			panic("ECS View Error: component cannot be both REQUIRED and EXCLUDED")
		}
		excludeMask = excludeMask.Set(id)
	}

	v.EntityIndex = entityIndex
	v.includeMask = includeMask
	v.compIDs = blueprint.CompIDs()
	v.excludeMask = excludeMask
}

func (v *View) Clear() {
	v.EntityIndex = nil
	v.includeMask = comp.Mask{}
	v.compIDs = nil
	v.excludeMask = comp.Mask{}
	v.BakedTablesCatalog.Clear()
}

func (v *View) BakeIfMatch(archetype *arch.Archetype) {
	if !archetype.Mask().IsEmpty() && archetype.Mask().Matches(v.includeMask, v.excludeMask) {
		v.BakedTablesCatalog.Add(archetype, v.compIDs)
	}
}

// All prepares the View for full chunk iteration and returns v.
// Call Next() to advance through matched entity chunks; read component slices
// with Slice[T]. The View holds iteration state directly — do not call All
// concurrently on the same View.
func (v *View) All() *View {
	v.mode = modeAll
	v.allIter = allIter{chunkIdx: -1}
	return v
}

// Filter prepares the View to iterate over selected entities and returns v.
// Call Next() to advance; read component pointers with At[T].
// The View holds iteration state directly — do not call Filter concurrently
// on the same View.
func (v *View) Filter(selected []uid.UID64) *View {
	v.mode = modeFilter
	v.filterIter = filterIter{selected: selected, lastArchID: arch.NullID}
	return v
}

// Next advances the iterator one step. Returns false when exhausted.
// The current mode (set by All or Filter) determines which iteration path runs.
func (v *View) Next() bool {
	switch v.mode {
	case modeAll:
		return v.nextAll()
	case modeFilter:
		return v.nextFilter()
	}
	return false
}
