package query

import (
	"time"
	"unsafe"

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
	modePaced
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
	verOffset  uintptr
	Entity     uid.UID64
	Idx        int
}

// seekBake caches a Seek target archetype's resolved offsets and presence.
type seekBake struct {
	offsets    []uintptr
	optOffsets []uintptr
	optPresent []bool
	baked      bool // this archetype has been resolved (offsets may legitimately be empty)
	complete   bool // it has a column for every tracked component
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
	pace        *comp.Pace
	// the run a walk began in, by the pace's step, and a paced walk's step
	now uint64
	d   time.Duration
	allIter
	filterIter
	seekTable      *colstore.Table
	seekBakes      []seekBake // by archetype, grown to the highest one Seek has met
	seekLastArchID arch.ID
	seekVerOffset  uintptr
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
	m.seekLastArchID = arch.NullID
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
	m.seekTable = nil
	m.seekBakes = nil
	m.seekLastArchID = arch.NullID
}

func (m *Matcher) BakeIfMatch(archetype *arch.Archetype) {
	if !archetype.Mask().IsEmpty() && archetype.Mask().Matches(m.includeMask, m.excludeMask) {
		m.BakedTablesCatalog.Add(archetype, m.compIDs, m.optCompIDs)
		m.BakedTables[len(m.BakedTables)-1].Period = m.pace.Period(archetype.Mask())
	}
}

// All starts full chunk iteration over matched archetypes; advance with Next.
func (m *Matcher) All() *Matcher {
	m.mode = modeAll
	m.allIter = allIter{chunkIdx: -1}
	m.seekLastArchID = arch.NullID // the Cursor's offsets will be a chunk's: Seek bakes afresh
	if m.pace != nil {
		m.now = m.pace.Step()
	}
	return m
}

// Paced starts the iteration All does over the archetypes whose turn it is this run, each over a
// step of d times its period; advance with Next.
func (m *Matcher) Paced(d time.Duration) *Matcher {
	m.All()
	m.mode, m.d = modePaced, d
	return m
}

// Due reports whether the archetype Next stands on has its turn this run.
func (m *Matcher) Due() bool { return m.due(&m.BakedTables[m.tableIdx]) }

// Dt is the step of the archetype a Paced iteration stands on.
func (m *Matcher) Dt() time.Duration {
	return m.d * time.Duration(m.BakedTables[m.tableIdx].Period)
}

func (m *Matcher) due(bt *BakedTable) bool {
	return bt.Period == 1 || bt.Period != 0 && m.now%bt.Period == bt.Period-1
}

// Pick starts iteration over the given entities, skipping those that do not
// match the mask; advance with Next.
func (m *Matcher) Pick(selected []uid.UID64) *Matcher {
	m.mode = modePick
	m.filterIter = filterIter{selected: selected, lastArchID: arch.NullID}
	m.seekLastArchID = arch.NullID
	return m
}

// Next advances the current All/Pick iteration; returns false when exhausted.
func (m *Matcher) Next() bool {
	switch m.mode {
	case modeAll:
		return m.nextAll()
	case modePick:
		return m.nextPick()
	case modePaced:
		return m.nextPaced()
	}
	return false
}

// Seek positions the Cursor at entID's storage slot, bypassing the query's
// include/exclude filters; returns false, leaving the Cursor untouched, if the
// entity does not exist or its archetype lacks a column for a tracked
// component. Consecutive Seeks into the same archetype reuse the cached table
// and column offsets.
func (m *Matcher) Seek(entID uid.UID64) bool {
	entry, ok := m.EntityIndex.Get(entID)
	if !ok {
		return false
	}
	if entry.ArchID != m.seekLastArchID {
		archetype := &m.archCatalog.Archetypes[entry.ArchID]
		if int(entry.ArchID) >= len(m.seekBakes) {
			m.growSeekBakes(int(entry.ArchID) + 1)
		}
		b := &m.seekBakes[entry.ArchID]
		if !b.baked {
			b.offsets, b.complete = archetype.Table.BakeOffsets(m.compIDs)
			b.optOffsets, b.optPresent = archetype.Table.BakeOptional(m.optCompIDs, archetype.Mask())
			b.baked = true
		}
		if !b.complete {
			return false
		}
		m.seekTable = &archetype.Table
		m.Cursor.Offsets = b.offsets
		m.Cursor.OptOffsets = b.optOffsets
		m.Cursor.OptPresent = b.optPresent
		m.seekLastArchID = entry.ArchID
		m.seekVerOffset = archetype.Table.VerOffset()
	}
	m.Cursor.Set(entry.ChunkPtr, uintptr(entry.Slot))
	m.Cursor.Ver = (*uint32)(unsafe.Add(entry.ChunkPtr, m.seekVerOffset))
	return true
}

// growSeekBakes makes room for the bakes of the archetypes below n, at least doubling: a new
// archetype met by Seek costs a copy once in a while, never a bake already made.
func (m *Matcher) growSeekBakes(n int) {
	grown := make([]seekBake, max(n, 2*len(m.seekBakes)))
	copy(grown, m.seekBakes)
	m.seekBakes = grown
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
		Ver:      *m.Cursor.Ver,
	}
}

// SeekH (Seek homogeneous) is Seek minus the alive and archetype-change
// checks: it assumes entID is alive and in the archetype cached by a prior
// Seek. Returns false when the archetype differs, leaving the Cursor exactly
// as it was — fall back to Seek. Undefined if entID is not alive.
func (m *Matcher) SeekH(entID uid.UID64) bool {
	entry := m.EntityIndex.GetUnchecked(entID)
	if entry.ArchID != m.seekLastArchID || m.seekLastArchID == arch.NullID {
		return false
	}
	m.Cursor.Set(entry.ChunkPtr, uintptr(entry.Slot))
	m.Cursor.Ver = (*uint32)(unsafe.Add(entry.ChunkPtr, m.seekVerOffset))
	return true
}
