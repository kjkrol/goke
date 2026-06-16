package query

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/soa"
)

// getPointer mirrors At: the row stride is supplied by the caller (in real code
// it is unsafe.Sizeof(T), a compile-time constant), not read from the BakedTable.
func getPointer(bt *BakedTable, pos soa.BlockPos, compIdx int, size uintptr) unsafe.Pointer {
	physPage := &bt.Table.Chunks[pos.ChunkIdx]
	return unsafe.Add(physPage.Ptr, bt.CompOffsets[compIdx]+(uintptr(pos.ChunkSlot)*size))
}

func getColumnStart(bt *BakedTable, chunk *soa.Chunk, compIdx int) unsafe.Pointer {
	return unsafe.Add(chunk.Ptr, bt.CompOffsets[compIdx])
}

func getEntityColumnStart(chunk *soa.Chunk) unsafe.Pointer {
	return chunk.Ptr
}

func TestBakedTable_GetPointer(t *testing.T) {
	data := make([]byte, 1024)
	chunk := soa.Chunk{Ptr: unsafe.Pointer(&data[0])}
	a := &arch.Archetype{}
	a.Table.Chunks = []soa.Chunk{chunk}

	bt := &BakedTable{
		Table:       &a.Table,
		CompOffsets: []uintptr{32, 64},
	}

	// compIdx 1: offset 64, size 16, slot 2 => 64 + (2 * 16) = 96
	ptr := getPointer(bt, soa.BlockPos{ChunkIdx: 0, ChunkSlot: 2}, 1, 16)
	expectedPtr := unsafe.Add(chunk.Ptr, 96)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}

func TestBakedTable_ColumnStarts(t *testing.T) {
	data := make([]byte, 1024)
	chunk := soa.Chunk{Ptr: unsafe.Pointer(&data[0])}

	bt := &BakedTable{
		CompOffsets: []uintptr{32, 64},
	}

	entPtr := getEntityColumnStart(&chunk)
	if entPtr != chunk.Ptr {
		t.Errorf("Expected EntityColumnStart %p, got %p", chunk.Ptr, entPtr)
	}

	comp0Ptr := getColumnStart(bt, &chunk, 0)
	expectedComp0Ptr := unsafe.Add(chunk.Ptr, 32)
	if comp0Ptr != expectedComp0Ptr {
		t.Errorf("Expected GetColumnStart(0) %p, got %p", expectedComp0Ptr, comp0Ptr)
	}

	comp1Ptr := getColumnStart(bt, &chunk, 1)
	expectedComp1Ptr := unsafe.Add(chunk.Ptr, 64)
	if comp1Ptr != expectedComp1Ptr {
		t.Errorf("Expected GetColumnStart(1) %p, got %p", expectedComp1Ptr, comp1Ptr)
	}
}
