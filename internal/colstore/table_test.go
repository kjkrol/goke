package colstore

import (
	"testing"

	"github.com/kjkrol/goke/internal/chunk"
	"github.com/kjkrol/goke/internal/comp"
)

func TestTable_LenTracking(t *testing.T) {
	compDefs := []comp.Def{
		{ID: 1, Size: 8, Align: 8},
	}

	var cs Table
	cs.Init(compDefs)

	if cs.Len() != 0 {
		t.Errorf("Expected initial Table.Len to be 0, got %d", cs.Len())
	}

	cs.chunkPack.AllocSlot()
	cs.chunkPack.AllocSlot()
	cs.chunkPack.AllocSlot()

	if cs.Len() != 3 {
		t.Errorf("Expected Table.Len to be 3 after 3 allocations, got %d", cs.Len())
	}

	if cs.chunkPack.ChunkLen(0) != 3 {
		t.Errorf("Expected chunk.Len to be 3, got %d", cs.chunkPack.ChunkLen(0))
	}

	cs.Clear()
	if cs.Len() != 0 {
		t.Errorf("Expected Table.Len to be 0 after Clear, got %d", cs.Len())
	}
}

func TestTable_ResolveTail_Reserved(t *testing.T) {
	compDefs := []comp.Def{
		{ID: 1, Size: 8, Align: 8},
	}

	var cs Table
	cs.Init(compDefs)

	cs.chunkPack.AddChunks(4)

	if cs.chunkPack.NumChunks() != 5 {
		t.Fatalf("Expected 5 pages initially, got %d", cs.chunkPack.NumChunks())
	}

	cs.chunkPack.Reserved = 0
	tailIdx, _ := cs.chunkPack.ResolveTail()
	if tailIdx != 0 {
		t.Errorf("Expected tailIdx 0, got %d", tailIdx)
	}
	if cs.chunkPack.NumChunks() != 1 {
		t.Errorf("Expected pages to be truncated to 1, got %d", cs.chunkPack.NumChunks())
	}

	cs.chunkPack.AddChunks(4)
	cs.chunkPack.Reserved = 2

	tailIdx, _ = cs.chunkPack.ResolveTail()

	if tailIdx != 0 {
		t.Errorf("Expected tailIdx 0 since no data exists, got %d", tailIdx)
	}
	if cs.chunkPack.NumChunks() != 3 {
		t.Errorf("Expected pages slice to be truncated to 3 (protecting reserved index 2), got %d", cs.chunkPack.NumChunks())
	}

	cs.chunkPack.AddChunks(2)
	cs.chunkPack.Extend(chunk.Idx(4), 1)
	cs.chunkPack.Reserved = 2

	tailIdx, _ = cs.chunkPack.ResolveTail()

	if tailIdx != 4 {
		t.Errorf("Expected tailIdx 4, got %d", tailIdx)
	}
	if cs.chunkPack.NumChunks() != 5 {
		t.Errorf("Expected pages to remain 5, got %d", cs.chunkPack.NumChunks())
	}
}
