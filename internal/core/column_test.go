package core

import (
	"testing"
	"unsafe"
)

func TestColumn_GetColumnStart(t *testing.T) {
	data := make([]byte, 1024)
	page := Page{Ptr: unsafe.Pointer(&data[0])}

	col := Column{
		PageOffset: 128,
	}

	ptr := col.GetColumnStart(&page)
	expectedPtr := unsafe.Add(page.Ptr, 128)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}
