package colstore

import (
	"unsafe"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/soa"
)

type Column struct {
	CompID   comp.ID
	CompSize uintptr
	Offset   uintptr // byte position of this column relative to the start of a Chunk
}

func (c *Column) At(chunk *soa.Chunk, pageSlot soa.ChunkSlot) unsafe.Pointer {
	return unsafe.Add(chunk.Ptr, c.Offset+uintptr(pageSlot)*c.CompSize)
}

func (c *Column) Base(chunk *soa.Chunk) unsafe.Pointer {
	return unsafe.Add(chunk.Ptr, c.Offset)
}
