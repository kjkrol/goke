package colstore

import (
	"unsafe"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/mem"
)

type Column struct {
	CompID   comp.ID
	CompSize uintptr
	Offset   uintptr // byte position of this column relative to the start of a Chunk
}

func (c *Column) At(chunkPtr unsafe.Pointer, slot mem.ChunkSlot) unsafe.Pointer {
	return unsafe.Add(chunkPtr, c.Offset+uintptr(slot)*c.CompSize)
}

func (c *Column) Base(chunkPtr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Add(chunkPtr, c.Offset)
}
