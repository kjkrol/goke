package colstore

import (
	"unsafe"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/soa"
)

type Column struct {
	CompID     comp.ID
	ItemSize   uintptr
	PageOffset uintptr
}

func (c *Column) At(chunk *soa.Chunk, pageSlot soa.ChunkSlot) unsafe.Pointer {
	return unsafe.Add(chunk.Ptr, c.PageOffset+uintptr(pageSlot)*c.ItemSize)
}

func (c *Column) Base(chunk *soa.Chunk) unsafe.Pointer {
	return unsafe.Add(chunk.Ptr, c.PageOffset)
}
