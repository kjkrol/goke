package query

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/soa"
)

func getPointer(ma *MatchedArch, pos soa.BlockPos, compIdx int) unsafe.Pointer {
	physPage := &ma.Table.Chunks[pos.ChunkIdx]
	return unsafe.Add(physPage.Ptr, ma.CompOffsets[compIdx]+(uintptr(pos.ChunkSlot)*ma.CompSizes[compIdx]))
}

func getColumnStart(ma *MatchedArch, chunk *soa.Chunk, compIdx int) unsafe.Pointer {
	return unsafe.Add(chunk.Ptr, ma.CompOffsets[compIdx])
}

func getEntityColumnStart(ma *MatchedArch, chunk *soa.Chunk) unsafe.Pointer {
	return unsafe.Add(chunk.Ptr, ma.EntityPageOffset)
}

func TestMatchedArch_GetPointer(t *testing.T) {
	data := make([]byte, 1024)
	chunk := soa.Chunk{Ptr: unsafe.Pointer(&data[0])}
	a := &arch.Archetype{}
	a.Table.Chunks = []soa.Chunk{chunk}

	ma := &MatchedArch{
		Table:       &a.Table,
		CompOffsets: []uintptr{32, 64},
		CompSizes:   []uintptr{8, 16},
	}

	// compIdx 1: offset 64, size 16, slot 2 => 64 + (2 * 16) = 96
	ptr := getPointer(ma, soa.BlockPos{ChunkIdx: 0, ChunkSlot: 2}, 1)
	expectedPtr := unsafe.Add(chunk.Ptr, 96)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}

func TestMatchedArch_ColumnStarts(t *testing.T) {
	data := make([]byte, 1024)
	chunk := soa.Chunk{Ptr: unsafe.Pointer(&data[0])}

	ma := &MatchedArch{
		EntityPageOffset: 16,
		CompOffsets:      []uintptr{32, 64},
	}

	entPtr := getEntityColumnStart(ma, &chunk)
	expectedEntPtr := unsafe.Add(chunk.Ptr, 16)
	if entPtr != expectedEntPtr {
		t.Errorf("Expected EntityColumnStart %p, got %p", expectedEntPtr, entPtr)
	}

	comp0Ptr := getColumnStart(ma, &chunk, 0)
	expectedComp0Ptr := unsafe.Add(chunk.Ptr, 32)
	if comp0Ptr != expectedComp0Ptr {
		t.Errorf("Expected GetColumnStart(0) %p, got %p", expectedComp0Ptr, comp0Ptr)
	}

	comp1Ptr := getColumnStart(ma, &chunk, 1)
	expectedComp1Ptr := unsafe.Add(chunk.Ptr, 64)
	if comp1Ptr != expectedComp1Ptr {
		t.Errorf("Expected GetColumnStart(1) %p, got %p", expectedComp1Ptr, comp1Ptr)
	}
}
