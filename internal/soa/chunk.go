package soa

import "unsafe"

type ChunkIdx uint32
type ChunkSlot uint32

// BlockPos identifies a slot within a Block by chunk index and slot within that chunk.
type BlockPos struct {
	ChunkIdx  ChunkIdx
	ChunkSlot ChunkSlot
}

type Chunk struct {
	data []byte
	Ptr  unsafe.Pointer
	Len  ChunkSlot
}

func (c *Chunk) GetPointer(offset uintptr, itemSize uintptr, slot ChunkSlot) unsafe.Pointer {
	return unsafe.Add(c.Ptr, offset+(uintptr(slot)*itemSize))
}

func NewChunk(chunkBytes uintptr) Chunk {
	data := make([]byte, chunkBytes)
	return Chunk{
		data: data,
		Ptr:  unsafe.Pointer(&data[0]),
		Len:  0,
	}
}

func newChunkSlice(n int, chunkBytes uintptr) []Chunk {
	bigBlock := make([]byte, uintptr(n)*chunkBytes)
	chunks := make([]Chunk, n)
	for i := range n {
		offset := uintptr(i) * chunkBytes
		chunks[i] = Chunk{
			data: bigBlock[offset : offset+chunkBytes : offset+chunkBytes],
			Ptr:  unsafe.Pointer(&bigBlock[offset]),
			Len:  0,
		}
	}
	return chunks
}

func alignUp(ptr, align uintptr) uintptr {
	return (ptr + align - 1) & ^(align - 1)
}
