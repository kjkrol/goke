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
	pace        *comp.Pace
	// Seek ignores a query's filters, so a filtered one bakes apart the archetypes holding every
	// tracked component; nil when they are just the matched ones
	seekable *seekTables
}

type seekTables struct {
	tracked comp.Mask
	BakedTablesCatalog
}

// rule is what a Matcher matches and the columns it bakes
type rule struct {
	include, exclude    comp.Mask
	compIDs, optCompIDs []comp.ID
	filtered            bool // Include or Exclude given: the Matcher keeps Seek tables apart
}

func newRule(s *comp.AccessSpec) rule {
	include := comp.NewMask(s)
	var exclude comp.Mask
	for _, id := range s.ExCompIDs {
		if include.IsSet(id) {
			panic("ECS Matcher Error: component cannot be both REQUIRED and EXCLUDED")
		}
		exclude = exclude.Set(id)
	}
	return rule{include: include, exclude: exclude, compIDs: s.CompIDs(),
		optCompIDs: s.OptCompIDs(), filtered: len(s.TagIDs) > 0 || len(s.ExCompIDs) > 0}
}

// wants reports whether an archetype of mask is one r matches or, r filtered, one Seek may need
func (r *rule) wants(mask comp.Mask) bool {
	return !mask.IsEmpty() && (r.filtered || mask.Matches(r.include, r.exclude))
}

func (m *Matcher) Init(entityIndex *addr.Index, archCatalog *arch.Catalog, accessSpec *comp.AccessSpec) {
	m.EntityIndex = entityIndex
	m.archCatalog = archCatalog
	m.pace = accessSpec.Pace
	if len(accessSpec.TagIDs) > 0 || len(accessSpec.ExCompIDs) > 0 {
		var tracked comp.Mask
		for _, def := range accessSpec.CompInfos {
			tracked = tracked.Set(def.ID)
		}
		m.seekable = &seekTables{tracked: tracked}
	}
}

func (m *Matcher) Clear() {
	m.EntityIndex = nil
	m.archCatalog = nil
	m.pace = nil
	m.BakedTablesCatalog.Clear()
	m.seekable = nil
}

// bake adds archetype to m's tables if r matches it, and to a filtered query's Seek tables if it
// holds every tracked component
func (m *Matcher) bake(archetype *arch.Archetype, r *rule) {
	mask := archetype.Mask()
	if mask.IsEmpty() {
		return
	}
	if mask.Matches(r.include, r.exclude) {
		m.BakedTablesCatalog.Add(archetype, r.compIDs, r.optCompIDs)
		m.BakedTables[len(m.BakedTables)-1].Period = m.pace.Period(mask)
	}
	if s := m.seekable; s != nil && mask.Matches(s.tracked, comp.Mask{}) {
		s.Add(archetype, r.compIDs, r.optCompIDs)
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
