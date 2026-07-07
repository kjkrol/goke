package chunk

import "unsafe"

type Idx uint32
type Slot uint32

// Pos identifies a slot within a Pack by chunk index and slot within that chunk.
type Pos struct {
	Idx  Idx
	Slot Slot
}

type Pack struct {
	chunks   []chunk
	Layout   Layout
	len      uint32
	Reserved Idx
	spare    []byte // backing array of the most recently trimmed chunk, reused by AddChunks
}

func (g *Pack) Init(layout Layout) {
	g.Layout = layout
	g.chunks = make([]chunk, 0, 16)
	g.AddChunks(1)
}

func (g *Pack) Len() uint32 { return g.len }

func (g *Pack) NumChunks() Idx { return Idx(len(g.chunks)) }

func (g *Pack) ChunkPtr(idx Idx) unsafe.Pointer {
	return g.chunks[idx].Ptr
}

func (g *Pack) ChunkLen(idx Idx) Slot {
	return g.chunks[idx].Len
}

// AllocSlot appends one slot; auto-allocates a new chunk if the current one is full.
func (g *Pack) AllocSlot() Pos {
	lastIdx := Idx(len(g.chunks) - 1)
	c := &g.chunks[lastIdx]

	if c.Len >= Slot(g.Layout.ChunkCap) {
		g.AddChunks(1)
		lastIdx++
		c = &g.chunks[lastIdx]
	}

	slot := c.Len
	c.Len++
	g.len++

	return Pos{Idx: lastIdx, Slot: slot}
}

// Extend advances chunk idx by n slots and returns the base pointer
// and the slot at which the new range starts.
func (g *Pack) Extend(idx Idx, n int) (base unsafe.Pointer, startSlot Slot) {
	startSlot = g.ChunkLen(idx)
	g.chunks[idx].Len += Slot(n)
	g.len += uint32(n)
	base = g.chunks[idx].Ptr
	return
}

// FreeSlot releases one slot from chunk idx.
func (g *Pack) FreeSlot(idx Idx) {
	g.chunks[idx].Len--
	g.len--
}

// FreeSlots releases n slots from the tail of chunk idx.
func (g *Pack) FreeSlots(idx Idx, n int) {
	g.chunks[idx].Len -= Slot(n)
	g.len -= uint32(n)
}

// NextNonEmptyChunk scans chunks starting at from and returns the first one
// with Len > 0. Returns the chunk index, its base pointer, its length, and
// whether a non-empty chunk was found.
func (g *Pack) NextNonEmptyChunk(from int) (idx int, ptr unsafe.Pointer, length int, ok bool) {
	for from < len(g.chunks) {
		if g.chunks[from].Len > 0 {
			return from, g.chunks[from].Ptr, int(g.chunks[from].Len), true
		}
		from++
	}
	return from, nil, 0, false
}

// AddChunks allocates n new chunks as one contiguous []byte and appends them.
// When n is 1 and a spare chunk (from a previous trim) is available, it is
// reused instead of allocating fresh memory.
func (g *Pack) AddChunks(n int) {
	start := len(g.chunks)

	if n == 1 && g.spare != nil {
		g.chunks = append(g.chunks, chunk{})
		g.chunks[start].init(g.spare)
		g.spare = nil
		return
	}

	bigBlock := make([]byte, uintptr(n)*g.Layout.ChunkBytes)
	g.chunks = append(g.chunks, make([]chunk, n)...)
	for i := range n {
		offset := uintptr(i) * g.Layout.ChunkBytes
		g.chunks[start+i].init(bigBlock[offset : offset+g.Layout.ChunkBytes : offset+g.Layout.ChunkBytes])
	}
}

// ReserveSlots ensures enough capacity for count slots starting from the current tail.
// Sets Reserved to prevent ResolveTail from trimming pre-allocated chunks.
// Returns the starting chunk index and the number of slots available in that first chunk.
func (g *Pack) ReserveSlots(count int) (startChunkIdx Idx, available int) {
	chunkIdx := Idx(len(g.chunks) - 1)
	available = int(g.Layout.ChunkCap) - int(g.chunks[chunkIdx].Len)

	if available == 0 {
		chunksNeeded := (count + int(g.Layout.ChunkCap) - 1) / int(g.Layout.ChunkCap)
		g.AddChunks(chunksNeeded)
		chunkIdx++
		available = int(g.Layout.ChunkCap)
	} else {
		chunksNeeded := (count - available + int(g.Layout.ChunkCap) - 1) / int(g.Layout.ChunkCap)
		if chunksNeeded > 0 {
			g.AddChunks(chunksNeeded)
		}
	}

	g.Reserved = Idx(len(g.chunks) - 1)
	return chunkIdx, available
}

// trimTrailing releases empty trailing chunks above the Reserved floor,
// stashing the last one released as the spare for AddChunks to reuse.
func (g *Pack) trimTrailing() {
	lastIdx := len(g.chunks) - 1
	floor := int(g.Reserved)

	for lastIdx > floor && g.chunks[lastIdx].Len == 0 {
		g.spare = g.chunks[lastIdx].data
		g.chunks = g.chunks[:lastIdx]
		lastIdx--
	}
}

// ResolveTail trims empty trailing chunks above the Reserved floor and
// returns the index and slot of the last occupied position.
func (g *Pack) ResolveTail() (Idx, Slot) {
	g.trimTrailing()

	tailIdx := len(g.chunks) - 1
	for tailIdx > 0 && g.chunks[tailIdx].Len == 0 {
		tailIdx--
	}

	tail := &g.chunks[tailIdx]
	return Idx(tailIdx), Slot(tail.Len - 1)
}

// SwapChunks exchanges the metadata (Ptr and Len) of chunks a and b. O(1).
// Used by Defragmenter to relocate a live chunk into a vacated slot without
// copying any entity data.
func (g *Pack) SwapChunks(a, b Idx) {
	g.chunks[a], g.chunks[b] = g.chunks[b], g.chunks[a]
}

// BulkFreeChunk marks chunk idx as empty: subtracts its Len from the total
// entity count and zeroes the Len field. The backing memory is retained until
// trimTrailing reclaims it. Call only when the chunk is being vacated as a
// unit (e.g. after SwapChunks moves its live data elsewhere).
func (g *Pack) BulkFreeChunk(idx Idx) {
	g.len -= uint32(g.chunks[idx].Len)
	g.chunks[idx].Len = 0
}

// ChunkIdxByPtr returns the index of the chunk whose Ptr equals ptr.
// Linear scan over chunks; intended for rare paths (Editor-style single-entity ops).
func (g *Pack) ChunkIdxByPtr(ptr unsafe.Pointer) Idx {
	for i, ch := range g.chunks {
		if ch.Ptr == ptr {
			return Idx(i)
		}
	}
	panic("chunk: ptr not found in pack")
}

func (g *Pack) Purge() {
	g.trimTrailing()
}

func (g *Pack) Clear() {
	for i := range g.chunks {
		g.chunks[i].Len = 0
	}
	g.chunks = g.chunks[:0]
	g.len = 0
}
