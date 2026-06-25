package addr

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/chunk"
)

func newPool() uid.UID64Pool {
	pool := uid.UID64Pool{}
	pool.Init(16, 16)
	return pool
}

func newIndex(cap int) Index {
	var idx Index
	idx.Init(cap)
	return idx
}

func TestIndex_UpsertAndGet(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, arch.ID(2), chunk.Pos{Idx: 0, Slot: 3})

	entry, ok := s.Get(e)
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if entry.ArchId != arch.ID(2) {
		t.Errorf("expected ArchId 2, got %d", entry.ArchId)
	}
	if entry.Pos.Idx != chunk.Idx(0) {
		t.Errorf("expected Idx 0, got %d", entry.Pos.Idx)
	}
	if entry.Pos.Slot != chunk.Slot(3) {
		t.Errorf("expected Slot 3, got %d", entry.Pos.Slot)
	}
}

func TestIndex_GetMissingEntity(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	_, ok := s.Get(e)
	if ok {
		t.Error("expected no entry for entity that was never stored")
	}
}

func TestIndex_GetStaleGeneration(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	old := pool.Next()
	pool.Release(old)
	current := pool.Next()

	s.Upsert(old, arch.ID(2), chunk.Pos{})

	_, ok := s.Get(current)
	if ok {
		t.Error("expected stale generation to not match")
	}
}

func TestIndex_Clear(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, arch.ID(3), chunk.Pos{})
	s.Clear(e)

	_, ok := s.Get(e)
	if ok {
		t.Error("expected entry to be gone after Clear")
	}
}

func TestIndex_ClearIgnoresStaleGeneration(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	old := pool.Next()
	pool.Release(old)
	current := pool.Next()

	s.Upsert(current, arch.ID(3), chunk.Pos{})
	s.Clear(old)

	_, ok := s.Get(current)
	if !ok {
		t.Error("Clear with stale generation should not remove current entry")
	}
}

func TestIndex_GrowsOnDemand(t *testing.T) {
	s := newIndex(2)
	var pool uid.UID64Pool
	pool.Init(64, 16)
	var e uid.UID64
	for range 11 {
		e = pool.Next()
	}

	s.Upsert(e, arch.ID(1), chunk.Pos{})

	entry, ok := s.Get(e)
	if !ok {
		t.Fatal("expected entry after grow")
	}
	if entry.ArchId != arch.ID(1) {
		t.Errorf("expected ArchId 1, got %d", entry.ArchId)
	}
}

func TestIndex_Reset(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, arch.ID(2), chunk.Pos{})
	s.Reset()

	_, ok := s.Get(e)
	if ok {
		t.Error("expected no entry after Reset")
	}
}

func TestIndex_Get_IndexBeyondEntries(t *testing.T) {
	s := newIndex(2)
	farID := uid.UID64(100) // index 100, generation 0 — never allocated, beyond len(entries)

	if _, ok := s.Get(farID); ok {
		t.Error("expected no entry for an index beyond the entries slice")
	}
}

func TestIndex_Clear_IndexBeyondEntries(t *testing.T) {
	s := newIndex(2)
	farID := uid.UID64(100)

	s.Clear(farID) // must not panic, and must not grow the slice
}

func TestIndex_EnsureCap(t *testing.T) {
	s := newIndex(2)

	s.EnsureCap(10)
	if cap(s.entries) < 10 {
		t.Errorf("expected cap >= 10 after EnsureCap, got %d", cap(s.entries))
	}

	before := cap(s.entries)
	s.EnsureCap(1) // already sufficient — must be a no-op
	if cap(s.entries) != before {
		t.Errorf("expected EnsureCap to leave cap unchanged when already sufficient, got %d (was %d)", cap(s.entries), before)
	}
}

func TestIndex_UpsertUnchecked(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.UpsertUnchecked(e, arch.ID(4), chunk.Pos{Idx: 1, Slot: 2})

	entry, ok := s.Get(e)
	if !ok {
		t.Fatal("expected entry after UpsertUnchecked")
	}
	if entry.ArchId != arch.ID(4) || entry.Pos.Idx != chunk.Idx(1) || entry.Pos.Slot != chunk.Slot(2) {
		t.Errorf("expected entry to match, got %+v", entry)
	}
}

func TestIndex_GetUnchecked(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, arch.ID(2), chunk.Pos{Idx: 0, Slot: 3})

	entry := s.GetUnchecked(e)
	if entry.ArchId != arch.ID(2) {
		t.Errorf("expected ArchId 2, got %d", entry.ArchId)
	}
	if entry.Pos.Idx != chunk.Idx(0) {
		t.Errorf("expected Idx 0, got %d", entry.Pos.Idx)
	}
	if entry.Pos.Slot != chunk.Slot(3) {
		t.Errorf("expected Slot 3, got %d", entry.Pos.Slot)
	}
}

func TestIndex_Upsert_Overwrite(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, arch.ID(1), chunk.Pos{Idx: 0, Slot: 0})
	s.Upsert(e, arch.ID(2), chunk.Pos{Idx: 1, Slot: 5})

	entry, ok := s.Get(e)
	if !ok {
		t.Fatal("expected entry after overwrite")
	}
	if entry.ArchId != arch.ID(2) || entry.Pos.Idx != chunk.Idx(1) || entry.Pos.Slot != chunk.Slot(5) {
		t.Errorf("expected overwritten entry, got %+v", entry)
	}
}
