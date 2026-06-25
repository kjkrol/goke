package query

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/addr"
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/colstore"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/iter"
)

type iterMode uint8

const (
	modeNone iterMode = iota
	modeAll
	modePick
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

// Matcher provides three ways to access entities matching a component mask:
// All (full chunk iteration), Pick (iterate a given entity subset), and Seek
// (position on a single known entity). Create once via NewMatcher/AddMatcher;
// the same instance can be reused across all three access patterns.
type Matcher struct {
	BakedTablesCatalog
	EntityIndex *addr.Index
	archCatalog *arch.Catalog
	includeMask comp.Mask
	compIDs     []comp.ID
	excludeMask comp.Mask
	mode        iterMode
	Cursor      iter.Cursor
	allIter
	filterIter
	seekTable      *colstore.Table
	seekOffsets    [arch.MaxID][]uintptr
	seekLastArchID arch.ID
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
	m.excludeMask = excludeMask
	m.seekLastArchID = arch.NullID
}

func (m *Matcher) Clear() {
	m.EntityIndex = nil
	m.archCatalog = nil
	m.includeMask = comp.Mask{}
	m.compIDs = nil
	m.excludeMask = comp.Mask{}
	m.BakedTablesCatalog.Clear()
	m.seekTable = nil
	m.seekOffsets = [arch.MaxID][]uintptr{}
	m.seekLastArchID = arch.NullID
}

func (m *Matcher) BakeIfMatch(archetype *arch.Archetype) {
	if !archetype.Mask().IsEmpty() && archetype.Mask().Matches(m.includeMask, m.excludeMask) {
		m.BakedTablesCatalog.Add(archetype, m.compIDs)
	}
}

// All prepares the Matcher for full chunk iteration and returns m.
// Call Next() to advance through matched entity chunks; read component slices
// with Slice[T]. The Matcher holds iteration state directly — do not call All
// concurrently on the same Matcher.
func (m *Matcher) All() *Matcher {
	m.mode = modeAll
	m.allIter = allIter{chunkIdx: -1}
	return m
}

// Pick prepares the Matcher to iterate over the given entities and returns m.
// Call Next() to advance; read component pointers with At[T]. Entities that do
// not match the Matcher's mask are skipped. The Matcher holds iteration state
// directly — do not call Pick concurrently on the same Matcher.
func (m *Matcher) Pick(selected []uid.UID64) *Matcher {
	m.mode = modePick
	m.filterIter = filterIter{selected: selected, lastArchID: arch.NullID}
	return m
}

// Next advances the iterator one step. Returns false when exhausted.
// The current mode (set by All or Pick) determines which iteration path runs.
func (m *Matcher) Next() bool {
	switch m.mode {
	case modeAll:
		return m.nextAll()
	case modePick:
		return m.nextPick()
	}
	return false
}

// Seek positions the Cursor at entID's storage slot, independent of the
// Matcher's include/exclude mask — the caller is trusted to know the entity
// carries the tracked components. Returns false if the entity does not exist
// or has been recycled.
//
// Consecutive Seeks that resolve to the same archetype reuse the cached table
// and column offsets, updating only the cursor's chunk and slot. Looping Seek
// over a set of entities from one archetype therefore resolves the archetype
// just once.
func (m *Matcher) Seek(entID uid.UID64) bool {
	entry, ok := m.EntityIndex.Get(entID)
	if !ok {
		return false
	}
	if entry.ArchId != m.seekLastArchID {
		m.seekTable = &m.archCatalog.Archetypes[entry.ArchId].Table
		offs := m.seekOffsets[entry.ArchId]
		if offs == nil {
			offs = m.seekTable.BakeOffsets(m.compIDs)
			m.seekOffsets[entry.ArchId] = offs
		}
		m.Cursor.Offsets = offs
		m.seekLastArchID = entry.ArchId
	}
	m.seekTable.PointCursor(&m.Cursor, entry.Pos)
	return true
}

// SeekH (Seek homogeneous) is Seek without the per-call archetype-change
// check and without the entity-alive check. It assumes entID is alive and
// shares the archetype already cached by a prior Seek call on this Matcher —
// call Seek once first to establish it, then switch to SeekH for the rest of
// a batch known to be alive and come from that one archetype.
//
// Returns false if entID's archetype turns out to differ from the cached
// one — the Cursor was positioned against the wrong table in that case and
// must not be used; call Seek for this entity instead. Behavior is
// undefined if entID is not alive.
func (m *Matcher) SeekH(entID uid.UID64) bool {
	entry := m.EntityIndex.GetUnchecked(entID)
	m.seekTable.PointCursor(&m.Cursor, entry.Pos)
	return entry.ArchId == m.seekLastArchID
}
