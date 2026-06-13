package soa

type Block struct {
	Chunks   []Chunk
	Layout   ChunkLayout
	Len      uint32
	Reserved ChunkIdx
}

func (b *Block) Init(layout ChunkLayout) {
	b.Layout = layout
	b.Chunks = make([]Chunk, 0, 16)
	b.addChunk()
}

func (b *Block) AllocSlot() (*Chunk, BlockPos) {
	lastIdx := ChunkIdx(len(b.Chunks) - 1)
	chunk := &b.Chunks[lastIdx]

	if chunk.Len >= ChunkSlot(b.Layout.ChunkCap) {
		b.addChunk()
		lastIdx++
		chunk = &b.Chunks[lastIdx]
	}

	slot := chunk.Len
	chunk.Len++
	b.Len++

	return chunk, BlockPos{ChunkIdx: lastIdx, ChunkSlot: slot}
}

func (b *Block) AddChunks(n int) {
	b.Chunks = append(b.Chunks, newChunkSlice(n, b.Layout.ChunkBytes)...)
}

func (b *Block) ResolveTail() (ChunkIdx, *Chunk) {
	lastIdx := len(b.Chunks) - 1
	floor := int(b.Reserved)

	for lastIdx > floor && b.Chunks[lastIdx].Len == 0 {
		b.Chunks = b.Chunks[:lastIdx]
		lastIdx--
	}

	tailIdx := lastIdx
	for tailIdx > 0 && b.Chunks[tailIdx].Len == 0 {
		tailIdx--
	}

	return ChunkIdx(tailIdx), &b.Chunks[tailIdx]
}

func (b *Block) GetChunk(idx ChunkIdx) *Chunk {
	return &b.Chunks[idx]
}

func (b *Block) Clear() {
	for i := range b.Chunks {
		b.Chunks[i].Len = 0
	}
	b.Chunks = b.Chunks[:0]
	b.Len = 0
}

func (b *Block) addChunk() {
	b.Chunks = append(b.Chunks, Chunk{})
	b.Chunks[len(b.Chunks)-1].InitChunk(b.Layout.ChunkBytes)
}
