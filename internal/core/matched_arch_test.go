package core

import (
	"testing"
	"unsafe"
)

func TestMatchedArch_GetPointer(t *testing.T) {
	data := make([]byte, 1024)
	page := Page{Ptr: unsafe.Pointer(&data[0])}
	arch := &Archetype{
		Memory: Memo{
			Pages: []Page{page},
		},
	}

	ma := &MatchedArch{
		Arch:        arch,
		CompOffsets: []uintptr{32, 64},
		CompSizes:   []uintptr{8, 16},
	}

	// Testing compIdx 1: offset 64, size 16, slot 2 => 64 + (2 * 16) = 96
	ptr := ma.GetPointer(0, 2, 1)
	expectedPtr := unsafe.Add(page.Ptr, 96)

	if ptr != expectedPtr {
		t.Errorf("Expected pointer %p, got %p", expectedPtr, ptr)
	}
}

func TestMatchedArch_ColumnStarts(t *testing.T) {
	data := make([]byte, 1024)
	page := Page{Ptr: unsafe.Pointer(&data[0])}

	ma := &MatchedArch{
		EntityPageOffset: 16,
		CompOffsets:      []uintptr{32, 64},
	}

	entPtr := ma.GetEntityColumnStart(&page)
	expectedEntPtr := unsafe.Add(page.Ptr, 16)
	if entPtr != expectedEntPtr {
		t.Errorf("Expected EntityColumnStart %p, got %p", expectedEntPtr, entPtr)
	}

	comp0Ptr := ma.GetColumnStart(&page, 0)
	expectedComp0Ptr := unsafe.Add(page.Ptr, 32)
	if comp0Ptr != expectedComp0Ptr {
		t.Errorf("Expected GetColumnStart(0) %p, got %p", expectedComp0Ptr, comp0Ptr)
	}

	comp1Ptr := ma.GetColumnStart(&page, 1)
	expectedComp1Ptr := unsafe.Add(page.Ptr, 64)
	if comp1Ptr != expectedComp1Ptr {
		t.Errorf("Expected GetColumnStart(1) %p, got %p", expectedComp1Ptr, comp1Ptr)
	}
}
