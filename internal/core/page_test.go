package core

import (
	"testing"
	"unsafe"
)

func TestPage_GetPointer(t *testing.T) {
	data := make([]byte, 1024)
	page := Page{Ptr: unsafe.Pointer(&data[0])}

	offset := uintptr(16)
	itemSize := uintptr(4)
	slot := PageSlot(3) // 16 + (3 * 4) = 28

	ptr := page.GetPointer(offset, itemSize, slot)
	expectedPtr := unsafe.Add(page.Ptr, 28)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}
