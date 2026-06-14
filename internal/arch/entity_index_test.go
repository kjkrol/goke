package arch

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/soa"
)

func newPool() uid.UID64Pool {
	return uid.NewUID64Pool(16, 16)
}

func TestEntityIndex_UpsertAndGet(t *testing.T) {
	s := NewEntityIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, ID(2), soa.BlockPos{ChunkIdx: 0, ChunkSlot: 3})

	link, ok := s.Get(e)
	if !ok {
		t.Fatal("expected link to be found")
	}
	if link.ArchId != ID(2) {
		t.Errorf("expected ArchId 2, got %d", link.ArchId)
	}
	if link.Pos.ChunkIdx != soa.ChunkIdx(0) {
		t.Errorf("expected ChunkIdx 0, got %d", link.Pos.ChunkIdx)
	}
	if link.Pos.ChunkSlot != soa.ChunkSlot(3) {
		t.Errorf("expected ChunkSlot 3, got %d", link.Pos.ChunkSlot)
	}
}

func TestEntityIndex_GetMissingEntity(t *testing.T) {
	s := NewEntityIndex(8)
	pool := newPool()
	e := pool.Next()

	_, ok := s.Get(e)
	if ok {
		t.Error("expected no link for entity that was never stored")
	}
}

func TestEntityIndex_GetStaleGeneration(t *testing.T) {
	s := NewEntityIndex(8)
	pool := newPool()
	old := pool.Next()
	pool.Release(old)
	current := pool.Next()

	s.Upsert(old, ID(2), soa.BlockPos{})

	_, ok := s.Get(current)
	if ok {
		t.Error("expected stale generation to not match")
	}
}

func TestEntityIndex_Clear(t *testing.T) {
	s := NewEntityIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, ID(3), soa.BlockPos{})
	s.Clear(e)

	_, ok := s.Get(e)
	if ok {
		t.Error("expected link to be gone after Clear")
	}
}

func TestEntityIndex_ClearIgnoresStaleGeneration(t *testing.T) {
	s := NewEntityIndex(8)
	pool := newPool()
	old := pool.Next()
	pool.Release(old)
	current := pool.Next()

	s.Upsert(current, ID(3), soa.BlockPos{})
	s.Clear(old)

	_, ok := s.Get(current)
	if !ok {
		t.Error("Clear with stale generation should not remove current link")
	}
}

func TestEntityIndex_GrowsOnDemand(t *testing.T) {
	s := NewEntityIndex(2)
	pool := uid.NewUID64Pool(64, 16)
	var e uid.UID64
	for range 11 {
		e = pool.Next()
	}

	s.Upsert(e, ID(1), soa.BlockPos{})

	link, ok := s.Get(e)
	if !ok {
		t.Fatal("expected link after grow")
	}
	if link.ArchId != ID(1) {
		t.Errorf("expected ArchId 1, got %d", link.ArchId)
	}
}

func TestEntityIndex_Reset(t *testing.T) {
	s := NewEntityIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, ID(2), soa.BlockPos{})
	s.Reset()

	_, ok := s.Get(e)
	if ok {
		t.Error("expected no link after Reset")
	}
}

func TestEntityIndex_Upsert_Overwrite(t *testing.T) {
	s := NewEntityIndex(8)
	pool := newPool()
	e := pool.Next()

	s.Upsert(e, ID(1), soa.BlockPos{ChunkIdx: 0, ChunkSlot: 0})
	s.Upsert(e, ID(2), soa.BlockPos{ChunkIdx: 1, ChunkSlot: 5})

	link, ok := s.Get(e)
	if !ok {
		t.Fatal("expected link after overwrite")
	}
	if link.ArchId != ID(2) || link.Pos.ChunkIdx != soa.ChunkIdx(1) || link.Pos.ChunkSlot != soa.ChunkSlot(5) {
		t.Errorf("expected overwritten link, got %+v", link)
	}
}
