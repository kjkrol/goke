package chunk

import "unsafe"

type Idx uint32
type Slot uint32

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

func (g *Pack) FreeSlot(idx Idx) {
	g.chunks[idx].Len--
	g.len--
}

// FreeSlots releases n slots from the tail of chunk idx.
func (g *Pack) FreeSlots(idx Idx, n int) {
	g.chunks[idx].Len -= Slot(n)
	g.len -= uint32(n)
}

// NextNonEmptyChunk returns the first chunk at or after from with Len > 0.
func (g *Pack) NextNonEmptyChunk(from int) (idx int, ptr unsafe.Pointer, length int, ok bool) {
	for from < len(g.chunks) {
		if g.chunks[from].Len > 0 {
			return from, g.chunks[from].Ptr, int(g.chunks[from].Len), true
		}
		from++
	}
	return from, nil, 0, false
}

// AddChunks appends n chunks backed by one contiguous allocation, reusing the
// spare chunk from a previous trim when n is 1.
func (g *Pack) AddChunks(n int) {
	start := len(g.chunks)

	if n == 1 && g.spare != nil {
		g.chunks = append(g.chunks, chunk{})
		g.chunks[start].init(g.spare)
		g.spare = nil
		return
	}

	var bigBlock []byte
	if g.Layout.NeedsScan {
		bigBlock = ScannableBytes(uintptr(n) * g.Layout.ChunkBytes)
	} else {
		bigBlock = make([]byte, uintptr(n)*g.Layout.ChunkBytes)
	}
	g.chunks = append(g.chunks, make([]chunk, n)...)
	for i := range n {
		offset := uintptr(i) * g.Layout.ChunkBytes
		g.chunks[start+i].init(bigBlock[offset : offset+g.Layout.ChunkBytes : offset+g.Layout.ChunkBytes])
	}
}

// ReserveSlots ensures capacity for count slots past the tail and sets Reserved
// so ResolveTail won't trim the pre-allocated chunks.
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

// SwapChunks exchanges the metadata (Ptr and Len) of chunks a and b — O(1),
// no entity bytes move.
func (g *Pack) SwapChunks(a, b Idx) {
	g.chunks[a], g.chunks[b] = g.chunks[b], g.chunks[a]
}

// BulkFreeChunk marks chunk idx empty, keeping its backing memory until
// trimTrailing reclaims it. Call only when the chunk is vacated as a unit.
func (g *Pack) BulkFreeChunk(idx Idx) {
	g.len -= uint32(g.chunks[idx].Len)
	g.chunks[idx].Len = 0
}

// ChunkIdxByPtr finds the chunk whose Ptr equals ptr — linear scan, rare paths only.
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
