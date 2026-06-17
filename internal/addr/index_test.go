package addr

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/mem"
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

	s.Upsert(e, arch.ID(2), mem.BlockPos{ChunkIdx: 0, ChunkSlot: 3})

	entry, ok := s.Get(e)
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if entry.ArchId != arch.ID(2) {
		t.Errorf("expected ArchId 2, got %d", entry.ArchId)
	}
	if entry.Pos.ChunkIdx != mem.ChunkIdx(0) {
		t.Errorf("expected ChunkIdx 0, got %d", entry.Pos.ChunkIdx)
	}
	if entry.Pos.ChunkSlot != mem.ChunkSlot(3) {
		t.Errorf("expected ChunkSlot 3, got %d", entry.Pos.ChunkSlot)
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

	s.Upsert(old, arch.ID(2), mem.BlockPos{})

	_, ok := s.Get(current)
	if ok {
		t.Error("expected stale generation to not match")
	}
}

func TestIndex_Clear(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, arch.ID(3), mem.BlockPos{})
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

	s.Upsert(current, arch.ID(3), mem.BlockPos{})
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

	s.Upsert(e, arch.ID(1), mem.BlockPos{})

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

	s.Upsert(e, arch.ID(2), mem.BlockPos{})
	s.Reset()

	_, ok := s.Get(e)
	if ok {
		t.Error("expected no entry after Reset")
	}
}

func TestIndex_Upsert_Overwrite(t *testing.T) {
	s := newIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, arch.ID(1), mem.BlockPos{ChunkIdx: 0, ChunkSlot: 0})
	s.Upsert(e, arch.ID(2), mem.BlockPos{ChunkIdx: 1, ChunkSlot: 5})

	entry, ok := s.Get(e)
	if !ok {
		t.Fatal("expected entry after overwrite")
	}
	if entry.ArchId != arch.ID(2) || entry.Pos.ChunkIdx != mem.ChunkIdx(1) || entry.Pos.ChunkSlot != mem.ChunkSlot(5) {
		t.Errorf("expected overwritten entry, got %+v", entry)
	}
}
