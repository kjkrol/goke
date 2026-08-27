package query

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/colstore"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
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

// seekBake caches a Seek target archetype's resolved offsets and presence.
type seekBake struct {
	offsets    []uintptr
	optOffsets []uintptr
	optPresent []bool
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
	optCompIDs  []comp.ID
	excludeMask comp.Mask
	mode        iterMode
	Cursor      iter.Cursor
	allIter
	filterIter
	seekTable      *colstore.Table
	seekBakes      [arch.MaxID]seekBake
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
	m.optCompIDs = accessSpec.OptCompIDs()
	m.excludeMask = excludeMask
	m.seekLastArchID = arch.NullID
}

func (m *Matcher) Clear() {
	m.EntityIndex = nil
	m.archCatalog = nil
	m.includeMask = comp.Mask{}
	m.compIDs = nil
	m.optCompIDs = nil
	m.excludeMask = comp.Mask{}
	m.BakedTablesCatalog.Clear()
	m.seekTable = nil
	m.seekBakes = [arch.MaxID]seekBake{}
	m.seekLastArchID = arch.NullID
}

func (m *Matcher) BakeIfMatch(archetype *arch.Archetype) {
	if !archetype.Mask().IsEmpty() && archetype.Mask().Matches(m.includeMask, m.excludeMask) {
		m.BakedTablesCatalog.Add(archetype, m.compIDs, m.optCompIDs)
	}
}

// All starts full chunk iteration over matched archetypes; advance with Next.
func (m *Matcher) All() *Matcher {
	m.mode = modeAll
	m.allIter = allIter{chunkIdx: -1}
	return m
}

// Pick starts iteration over the given entities, skipping those that do not
// match the mask; advance with Next.
func (m *Matcher) Pick(selected []uid.UID64) *Matcher {
	m.mode = modePick
	m.filterIter = filterIter{selected: selected, lastArchID: arch.NullID}
	return m
}

// Next advances the current All/Pick iteration; returns false when exhausted.
func (m *Matcher) Next() bool {
	switch m.mode {
	case modeAll:
		return m.nextAll()
	case modePick:
		return m.nextPick()
	}
	return false
}

// Seek positions the Cursor at entID's storage slot, bypassing the mask;
// returns false if the entity does not exist. Consecutive Seeks into the
// same archetype reuse the cached table and column offsets.
func (m *Matcher) Seek(entID uid.UID64) bool {
	entry, ok := m.EntityIndex.Get(entID)
	if !ok {
		return false
	}
	if entry.ArchID != m.seekLastArchID {
		m.seekTable = &m.archCatalog.Archetypes[entry.ArchID].Table
		b := &m.seekBakes[entry.ArchID]
		if b.offsets == nil {
			b.offsets = m.seekTable.BakeOffsets(m.compIDs)
			b.optOffsets, b.optPresent = m.seekTable.BakeOptional(m.optCompIDs)
		}
		m.Cursor.Offsets = b.offsets
		m.Cursor.OptOffsets = b.optOffsets
		m.Cursor.OptPresent = b.optPresent
		m.seekLastArchID = entry.ArchID
	}
	m.Cursor.Set(entry.ChunkPtr, uintptr(entry.Slot))
	return true
}

// ChunkSnapshot captures the chunk most recently advanced to by Next in All
// mode, for use with CmdBuf.MassMigrate.
func (m *Matcher) ChunkSnapshot() bulk.ChunkSnapshot {
	bt := &m.BakedTables[m.tableIdx]
	return bulk.ChunkSnapshot{
		ArchID:   bt.ArchID,
		ChunkPtr: m.Cursor.Base,
		ChunkIdx: colstore.Idx(m.chunkIdx),
		TableVer: bt.Table.Version(),
	}
}

// SeekH (Seek homogeneous) is Seek minus the alive and archetype-change
// checks: it assumes entID is alive and in the archetype cached by a prior
// Seek. Returns false when the archetype differs — the Cursor is then
// invalid; fall back to Seek. Undefined if entID is not alive.
func (m *Matcher) SeekH(entID uid.UID64) bool {
	entry := m.EntityIndex.GetUnchecked(entID)
	m.Cursor.Set(entry.ChunkPtr, uintptr(entry.Slot))
	return entry.ArchID == m.seekLastArchID
}
