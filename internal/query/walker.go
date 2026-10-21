package query

import (
	"time"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/colstore"
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
	Entity     uid.UID64
	Idx        int
}

// Walker is one walk over a Matcher's archetypes: its Cursor and the place All, Pick, Paced and
// Seek have reached. Walkers of one Matcher may walk on as many goroutines while no entity or
// archetype changes.
type Walker struct {
	m           *Matcher
	entityIndex *addr.Index
	archCatalog *arch.Catalog
	// the Matcher's tables as a walk began: archetypes come only at a Sync, never during one
	tables BakedTablesCatalog
	mode   iterMode
	Cursor iter.Cursor
	// the run a walk began in, by the pace's step, and a paced walk's step
	now uint64
	d   time.Duration
	allIter
	filterIter
	seekLastArchID arch.ID
}

// Init has w walk m, standing nowhere.
func (w *Walker) Init(m *Matcher) {
	*w = Walker{m: m, entityIndex: m.EntityIndex, archCatalog: m.archCatalog,
		seekLastArchID: arch.NullID}
}

// Matcher is what w walks.
func (w *Walker) Matcher() *Matcher { return w.m }

// Clear ends w's walk: Next reports nothing till All, Pick or Paced starts another.
func (w *Walker) Clear() { w.Init(w.m) }

// All starts full chunk iteration over matched archetypes; advance with Next.
func (w *Walker) All() *Walker {
	w.mode = modeAll
	w.tables = w.m.BakedTablesCatalog
	w.allIter = allIter{chunkIdx: -1}
	w.seekLastArchID = arch.NullID // the Cursor's offsets will be a chunk's: Seek bakes afresh
	if w.m.pace != nil {
		w.now = w.m.pace.Step()
	}
	return w
}

// Paced starts the iteration All does over the archetypes whose turn it is this run, each over a
// step of d times its period; advance with Next.
func (w *Walker) Paced(d time.Duration) *Walker {
	w.All()
	w.mode, w.d = modePaced, d
	return w
}

// Due reports whether the archetype Next stands on has its turn this run.
func (w *Walker) Due() bool { return w.due(&w.tables.BakedTables[w.tableIdx]) }

// Dt is the step of the archetype a Paced iteration stands on.
func (w *Walker) Dt() time.Duration {
	return w.d * time.Duration(w.tables.BakedTables[w.tableIdx].Period)
}

func (w *Walker) due(bt *BakedTable) bool {
	return bt.Period == 1 || bt.Period != 0 && w.now%bt.Period == bt.Period-1
}

// Pick starts iteration over the given entities, skipping those that do not
// match the mask; advance with Next.
func (w *Walker) Pick(selected []uid.UID64) *Walker {
	w.mode = modePick
	w.tables = w.m.BakedTablesCatalog
	w.filterIter = filterIter{selected: selected, lastArchID: arch.NullID}
	w.seekLastArchID = arch.NullID
	return w
}

// Next advances the current All/Pick iteration; returns false when exhausted.
func (w *Walker) Next() bool {
	switch w.mode {
	case modeAll:
		return w.nextAll()
	case modePick:
		return w.nextPick()
	case modePaced:
		return w.nextPaced()
	}
	return false
}

// Seek positions the Cursor at entID's storage slot, bypassing the query's
// include/exclude filters; returns false, leaving the Cursor untouched, if the
// entity does not exist or its archetype lacks a column for a tracked
// component. Consecutive Seeks into the same archetype reuse the cached
// column offsets.
func (w *Walker) Seek(entID uid.UID64) bool {
	entry, ok := w.entityIndex.Get(entID)
	if !ok {
		return false
	}
	if id := entry.ArchID; id != w.seekLastArchID {
		m := w.m
		var b *seekBake
		if int(id) < len(m.seekAt) && m.seekAt[id] != 0 {
			b = &m.seekBakes[m.seekAt[id]-1]
		} else {
			b = m.bake(id)
		}
		if !b.complete {
			return false
		}
		w.Cursor.Offsets = b.offsets
		w.Cursor.OptOffsets = b.optOffsets
		w.Cursor.OptPresent = b.optPresent
		w.seekLastArchID = id
	}
	w.Cursor.Set(entry.ChunkPtr, uintptr(entry.Slot))
	return true
}

// SeekH (Seek homogeneous) is Seek minus the alive and archetype-change
// checks: it assumes entID is alive and in the archetype cached by a prior
// Seek. Returns false when the archetype differs, leaving the Cursor exactly
// as it was — fall back to Seek. Undefined if entID is not alive.
func (w *Walker) SeekH(entID uid.UID64) bool {
	entry := w.entityIndex.GetUnchecked(entID)
	if entry.ArchID != w.seekLastArchID || w.seekLastArchID == arch.NullID {
		return false
	}
	w.Cursor.Set(entry.ChunkPtr, uintptr(entry.Slot))
	return true
}

// ChunkSnapshot captures the chunk most recently advanced to by Next in All
// mode, for use with CmdBuf.MassMigrate.
func (w *Walker) ChunkSnapshot() bulk.ChunkSnapshot {
	bt := &w.tables.BakedTables[w.tableIdx]
	return bulk.ChunkSnapshot{
		ArchID:   bt.ArchID,
		ChunkPtr: w.Cursor.Base,
		ChunkIdx: colstore.Idx(w.chunkIdx),
		TableVer: bt.Table.Version(),
	}
}
