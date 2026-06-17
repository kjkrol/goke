package mem

import "unsafe"

type ChunkIdx uint32
type ChunkSlot uint32

// BlockPos identifies a slot within a Block by chunk index and slot within that chunk.
type BlockPos struct {
	ChunkIdx  ChunkIdx
	ChunkSlot ChunkSlot
}

type Block struct {
	chunks   []chunk
	Layout   ChunkLayout
	len      uint32
	Reserved ChunkIdx
}

func (b *Block) Len() uint32 { return b.len }

func (b *Block) Init(layout ChunkLayout) {
	b.Layout = layout
	b.chunks = make([]chunk, 0, 16)
	b.AddChunks(1)
}

func (b *Block) NumChunks() int {
	return len(b.chunks)
}

func (b *Block) ChunkPtr(idx ChunkIdx) unsafe.Pointer {
	return b.chunks[idx].Ptr
}

func (b *Block) ChunkLen(idx ChunkIdx) ChunkSlot {
	return b.chunks[idx].Len
}

func (b *Block) AllocSlot() BlockPos {
	lastIdx := ChunkIdx(len(b.chunks) - 1)
	c := &b.chunks[lastIdx]

	if c.Len >= ChunkSlot(b.Layout.ChunkCap) {
		b.AddChunks(1)
		lastIdx++
		c = &b.chunks[lastIdx]
	}

	slot := c.Len
	c.Len++
	b.len++

	return BlockPos{ChunkIdx: lastIdx, ChunkSlot: slot}
}

func (b *Block) AllocSlots(idx ChunkIdx, n int) {
	b.chunks[idx].Len += ChunkSlot(n)
	b.len += uint32(n)
}

func (b *Block) FreeSlot(idx ChunkIdx) {
	b.chunks[idx].Len--
	b.len--
}

// NextNonEmptyChunk scans chunks starting at from and returns the first one
// with Len > 0. Returns the chunk index, its base pointer, its length, and
// whether a non-empty chunk was found.
func (b *Block) NextNonEmptyChunk(from int) (idx int, ptr unsafe.Pointer, length int, ok bool) {
	for from < len(b.chunks) {
		if b.chunks[from].Len > 0 {
			return from, b.chunks[from].Ptr, int(b.chunks[from].Len), true
		}
		from++
	}
	return from, nil, 0, false
}

func (b *Block) AddChunks(n int) {
	bigBlock := make([]byte, uintptr(n)*b.Layout.ChunkBytes)
	start := len(b.chunks)
	b.chunks = append(b.chunks, make([]chunk, n)...)
	for i := range n {
		offset := uintptr(i) * b.Layout.ChunkBytes
		b.chunks[start+i].init(bigBlock[offset : offset+b.Layout.ChunkBytes : offset+b.Layout.ChunkBytes])
	}
}

func (b *Block) PrepareSlots(count int) (startChunkIdx ChunkIdx, available int) {
	chunkIdx := ChunkIdx(b.NumChunks() - 1)
	available = int(b.Layout.ChunkCap) - int(b.ChunkLen(chunkIdx))

	if available == 0 {
		chunksNeeded := (count + int(b.Layout.ChunkCap) - 1) / int(b.Layout.ChunkCap)
		b.AddChunks(chunksNeeded)
		chunkIdx++
		available = int(b.Layout.ChunkCap)
	} else {
		chunksNeeded := (count - available + int(b.Layout.ChunkCap) - 1) / int(b.Layout.ChunkCap)
		if chunksNeeded > 0 {
			b.AddChunks(chunksNeeded)
		}
	}

	b.Reserved = ChunkIdx(b.NumChunks() - 1)
	return chunkIdx, available
}

func (b *Block) ResolveTail() (ChunkIdx, ChunkSlot) {
	lastIdx := len(b.chunks) - 1
	floor := int(b.Reserved)

	for lastIdx > floor && b.chunks[lastIdx].Len == 0 {
		b.chunks = b.chunks[:lastIdx]
		lastIdx--
	}

	tailIdx := lastIdx
	for tailIdx > 0 && b.chunks[tailIdx].Len == 0 {
		tailIdx--
	}

	tail := &b.chunks[tailIdx]
	return ChunkIdx(tailIdx), ChunkSlot(tail.Len - 1)
}

func (b *Block) Clear() {
	for i := range b.chunks {
		b.chunks[i].Len = 0
	}
	b.chunks = b.chunks[:0]
	b.len = 0
}
