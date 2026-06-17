package mem

import "unsafe"

type chunk struct {
	data []byte
	Ptr  unsafe.Pointer
	Len  ChunkSlot
}

func (c *chunk) init(data []byte) {
	c.data = data
	c.Ptr = unsafe.Pointer(&data[0])
	c.Len = 0
}
