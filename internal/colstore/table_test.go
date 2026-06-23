package colstore

import (
	"testing"

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
