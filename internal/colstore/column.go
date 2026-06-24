package colstore

import (
	"unsafe"

	"github.com/kjkrol/goke/v2/internal/chunk"
	"github.com/kjkrol/goke/v2/internal/comp"
)

type ColDef struct {
	CompID   comp.ID
	CompSize uintptr
	Offset   uintptr // byte position of this column relative to the start of a chunk
}

func (c *ColDef) At(chunkPtr unsafe.Pointer, slot chunk.Slot) unsafe.Pointer {
	return unsafe.Add(chunkPtr, c.Offset+uintptr(slot)*c.CompSize)
}

func (c *ColDef) Base(chunkPtr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Add(chunkPtr, c.Offset)
}
