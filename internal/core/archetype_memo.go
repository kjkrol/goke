package core

import "unsafe"

const PageSize = 16 * 1024 // 16KB - fits L1 Cache

type ChunkIdx uint32 // Index of the chunk in Memo.Pages slice
// ChunkRow is a type alias for the index within a chunk.
// Using uint32 ensures alignment and supports >255 entities per chunk.
type ChunkRow uint32

//------------------------------------------------------------------------------
//                          Memo (Memory Manager)
//------------------------------------------------------------------------------

type Memo struct {
	// Pages holds pointers to all allocated chunks.
	// Using a slice allows O(1) access by ChunkIdx, which is crucial for EntityLinkStore.
	Pages  []*chunk
	Layout ChunkLayout
	Len    uint32 // Global entity count (optional, but useful)
}

func (b *Memo) Init(compInfos []ComponentInfo) {
	b.Layout = CalculateLayout(compInfos)

	// Pre-allocate slice capacity to avoid frequent resizing at start
	b.Pages = make([]*chunk, 0, 16)
	b.Len = 0

	b.addChunk()
}

// AllocSlot allocates space for a new entity.
// It returns:
// 1. *chunk  -> Pointer for immediate data writing (fastest access)
// 2. uint32  -> ChunkIdx (to store in EntityLinkStore)
// 3. ChunkRow -> Row index within the chunk (to store in EntityLinkStore)
func (b *Memo) AllocSlot() (*chunk, ChunkIdx, ChunkRow) {
	lastIdx := ChunkIdx(len(b.Pages) - 1)
	c := b.Pages[lastIdx]

	// If the current chunk is full, create a new one
	if c.Len >= ChunkRow(b.Layout.ChunkCap) {
		b.addChunk()
		lastIdx++
		c = b.Pages[lastIdx]
	}

	row := c.Len
	c.Len++
	b.Len++

	return c, lastIdx, row
}

// GetChunk provides O(1) access to a chunk by its index.
// Used when moving/removing entities based on EntityLinkStore data.
func (b *Memo) GetChunk(idx ChunkIdx) *chunk {
	// In production, you might skip bounds check if you trust LinkStore
	return b.Pages[idx]
}

func (b *Memo) addChunk() {
	data := make([]byte, PageSize)

	newChunk := &chunk{
		data: data,
		ptr:  unsafe.Pointer(&data[0]),
		Len:  0,
	}

	b.Pages = append(b.Pages, newChunk)
}

func (b *Memo) Clear() {
	// 1. Zero out memory for GC safety
	for _, c := range b.Pages {
		clear(c.data)
		c.Len = 0
	}

	// 2. Reset the slice
	// We can keep the underlying array capacity to avoid re-allocations on restart
	b.Pages = b.Pages[:0]
	b.Len = 0

	// 3. Immediately add the first fresh chunk
	b.addChunk()
}

//------------------------------------------------------------------------------
//                          chunk
//------------------------------------------------------------------------------

type chunk struct {
	data []byte
	ptr  unsafe.Pointer
	Len  ChunkRow
}

//------------------------------------------------------------------------------
//                          CalculateLayout
//------------------------------------------------------------------------------

type ChunkLayout struct {
	ChunkCap uint32
	Offsets  []uintptr
}

// CalculateLayout computes the optimal memory layout for a chunk.
func CalculateLayout(compInfos []ComponentInfo) ChunkLayout {

	totalStride := unsafe.Sizeof(Entity(0))
	for _, info := range compInfos {
		totalStride += info.Size
	}

	capacity := uintptr(PageSize) / totalStride
	if capacity == 0 {
		panic("Entity layout too large for a single memory page (16KB)")
	}

	for capacity > 0 {
		offsets := make([]uintptr, len(compInfos)+1)
		currentOffset := uintptr(0)
		fits := true

		// --- STEP A: Entity ID ---
		entityAlign := unsafe.Alignof(Entity(0))
		currentOffset = alignUp(currentOffset, entityAlign)
		offsets[0] = currentOffset
		currentOffset += unsafe.Sizeof(Entity(0)) * capacity

		// --- STEP B: Components ---
		for i, info := range compInfos {
			currentOffset = alignUp(currentOffset, info.Align)
			offsets[i+1] = currentOffset
			currentOffset += info.Size * capacity

			if currentOffset > PageSize {
				fits = false
				break
			}
		}

		if fits {
			return ChunkLayout{
				ChunkCap: uint32(capacity),
				Offsets:  offsets,
			}
		}

		capacity--
	}

	panic("Components too large for PageSize")
}

func alignUp(ptr, align uintptr) uintptr {
	return (ptr + align - 1) & ^(align - 1)
}

// -----------------------------------------------------------------------------
// Low-Level Helpers
// -----------------------------------------------------------------------------

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func zeroMemory(ptr unsafe.Pointer, size uintptr) {
	clear(unsafe.Slice((*byte)(ptr), size))
}
