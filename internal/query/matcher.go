package query

import (
	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/comp"
)

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
	// Seek ignores a query's filters, so a filtered one bakes apart the archetypes holding every
	// tracked component; nil when they are just the matched ones
	seekable *seekTables
}

type seekTables struct {
	tracked comp.Mask
	BakedTablesCatalog
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
	if len(accessSpec.TagIDs) > 0 || len(accessSpec.ExCompIDs) > 0 {
		var tracked comp.Mask
		for _, id := range m.compIDs {
			tracked = tracked.Set(id)
		}
		m.seekable = &seekTables{tracked: tracked}
	}
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
	m.seekable = nil
}

func (m *Matcher) BakeIfMatch(archetype *arch.Archetype) {
	mask := archetype.Mask()
	if mask.IsEmpty() {
		return
	}
	if mask.Matches(m.includeMask, m.excludeMask) {
		m.BakedTablesCatalog.Add(archetype, m.compIDs, m.optCompIDs)
		m.BakedTables[len(m.BakedTables)-1].Period = m.pace.Period(mask)
	}
	if s := m.seekable; s != nil && mask.Matches(s.tracked, comp.Mask{}) {
		s.Add(archetype, m.compIDs, m.optCompIDs)
	}
}

// seekTable is the table Seek positions by in archetype id; nil if it has no column for a
// tracked component
func (m *Matcher) seekTable(id arch.ID) *BakedTable {
	tables := &m.BakedTablesCatalog
	if m.seekable != nil {
		tables = &m.seekable.BakedTablesCatalog
	}
	if bt := tables.Get(id); bt != nil && bt.Complete {
		return bt
	}
	return nil
}
