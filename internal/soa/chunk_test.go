package soa

import (
	"testing"
	"unsafe"
)

func TestChunk_GetPointer(t *testing.T) {
	data := make([]byte, 1024)
	chunk := Chunk{Ptr: unsafe.Pointer(&data[0])}

	offset := uintptr(16)
	itemSize := uintptr(4)
	slot := ChunkSlot(3) // 16 + (3 * 4) = 28

	ptr := chunk.GetPointer(offset, itemSize, slot)
	expectedPtr := unsafe.Add(chunk.Ptr, 28)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}
