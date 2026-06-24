package colstore

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v2/internal/chunk"
)

func TestColumn_Base(t *testing.T) {
	data := make([]byte, 1024)
	chunkPtr := unsafe.Pointer(&data[0])

	col := ColDef{
		Offset: 128,
	}

	ptr := col.Base(chunkPtr)
	expectedPtr := unsafe.Add(chunkPtr, 128)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}

func TestColumn_At(t *testing.T) {
	data := make([]byte, 1024)
	chunkPtr := unsafe.Pointer(&data[0])

	col := ColDef{
		Offset:   64,
		CompSize: 8,
	}

	// slot 3: 64 + 3*8 = 88
	ptr := col.At(chunkPtr, chunk.Slot(3))
	expectedPtr := unsafe.Add(chunkPtr, 88)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}
