package colstore

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/internal/soa"
)

func TestColumn_Base(t *testing.T) {
	data := make([]byte, 1024)
	chunk := soa.Chunk{Ptr: unsafe.Pointer(&data[0])}

	col := Column{
		PageOffset: 128,
	}

	ptr := col.Base(&chunk)
	expectedPtr := unsafe.Add(chunk.Ptr, 128)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}
