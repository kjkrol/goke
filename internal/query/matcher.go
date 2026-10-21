package query

import (
	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/comp"
)

// seekBake caches a Seek target archetype's resolved offsets and presence.
type seekBake struct {
	offsets    []uintptr
	optOffsets []uintptr
	optPresent []bool
	complete   bool // it has a column for every tracked component
}

// Matcher is what a query matches: the archetypes of a component mask, baked into tables as they
// come. Walkers walk it — All, Pick, Paced, Seek — each with a cursor of its own. Create once via
// NewMatcher/AddMatcher.
type Matcher struct {
	BakedTablesCatalog
	EntityIndex *addr.Index
	archCatalog *arch.Catalog
	includeMask comp.Mask
	compIDs     []comp.ID
	optCompIDs  []comp.ID
	excludeMask comp.Mask
	pace        *comp.Pace
	seekAt      []uint16   // by archetype: 0 none yet, else 1 + the place of its bake
	seekBakes   []seekBake // the archetypes Seek has met, in the order it met them
}

func (m *Matcher) Init(entityIndex *addr.Index, archCatalog *arch.Catalog, accessSpec *comp.AccessSpec) {
	includeMask := comp.NewMask(accessSpec)

	var excludeMask comp.Mask
	for _, id := range accessSpec.ExCompIDs {
		if includeMask.IsSet(id) {
			panic("ECS Matcher Error: component cannot be both REQUIRED and EXCLUDED")
		}
		excludeMask = excludeMask.Set(id)
	}

	m.EntityIndex = entityIndex
	m.archCatalog = archCatalog
	m.includeMask = includeMask
	m.compIDs = accessSpec.CompIDs()
	m.optCompIDs = accessSpec.OptCompIDs()
	m.excludeMask = excludeMask
	m.pace = accessSpec.Pace
}

func (m *Matcher) Clear() {
	m.EntityIndex = nil
	m.archCatalog = nil
	m.includeMask = comp.Mask{}
	m.compIDs = nil
	m.optCompIDs = nil
	m.excludeMask = comp.Mask{}
	m.pace = nil
	m.BakedTablesCatalog.Clear()
	m.seekAt, m.seekBakes = nil, nil
}

func (m *Matcher) BakeIfMatch(archetype *arch.Archetype) {
	if !archetype.Mask().IsEmpty() && archetype.Mask().Matches(m.includeMask, m.excludeMask) {
		m.BakedTablesCatalog.Add(archetype, m.compIDs, m.optCompIDs)
		m.BakedTables[len(m.BakedTables)-1].Period = m.pace.Period(archetype.Mask())
	}
}

// bake works out the offsets of m's columns in archetype id at the first Seek there, making
// room in seekAt for id as append does: at least doubling, 2 bytes a slot.
func (m *Matcher) bake(id arch.ID) *seekBake {
	if int(id) >= len(m.seekAt) {
		grown := make([]uint16, max(int(id)+1, 2*len(m.seekAt)))
		copy(grown, m.seekAt)
		m.seekAt = grown
	}
	a := &m.archCatalog.Archetypes[id]
	var b seekBake
	b.offsets, b.complete = a.Table.BakeOffsets(m.compIDs)
	b.optOffsets, b.optPresent = a.Table.BakeOptional(m.optCompIDs, a.Mask())
	m.seekBakes = append(m.seekBakes, b)
	m.seekAt[id] = uint16(len(m.seekBakes))
	return &m.seekBakes[len(m.seekBakes)-1]
}
