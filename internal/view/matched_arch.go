package view

import (
	"unsafe"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/mem"
)

type MatchedArch struct {
	Arch             *arch.Archetype
	EntityPageOffset uintptr
	CompOffsets      []uintptr
	CompSizes        []uintptr
}

func (ma *MatchedArch) Clear() {
	ma.Arch = nil
	clear(ma.CompOffsets)
	ma.CompOffsets = nil
	clear(ma.CompSizes)
	ma.CompSizes = nil
	ma.EntityPageOffset = 0
}

func (ma *MatchedArch) GetPointer(pageIdx mem.PageIdx, slot mem.PageSlot, compIdx int) unsafe.Pointer {
	physPage := &ma.Arch.Memory.Pages[pageIdx]
	return unsafe.Add(physPage.Ptr, ma.CompOffsets[compIdx]+(uintptr(slot)*ma.CompSizes[compIdx]))
}

func (ma *MatchedArch) GetColumnStart(page *mem.Page, compIdx int) unsafe.Pointer {
	return unsafe.Add(page.Ptr, ma.CompOffsets[compIdx])
}

func (ma *MatchedArch) GetEntityColumnStart(page *mem.Page) unsafe.Pointer {
	return unsafe.Add(page.Ptr, ma.EntityPageOffset)
}
