package colstore

import (
	"sort"
	"unsafe"

	"github.com/kjkrol/uid"
)

// SlotRef identifies an entity slot by chunk memory pointer and slot; valid
// across SwapChunks, unlike Pos. Idx is the chunk's Pack position at capture time.
type SlotRef struct {
	Ptr  unsafe.Pointer
	Idx  Idx
	Slot Slot
}

// SlotMove records an entity relocated within a table during compaction.
type SlotMove struct {
	ID      uid.UID64
	NewPtr  unsafe.Pointer
	NewSlot Slot
}

// CompactScratch holds Defragmenter's working buffers; grows lazily, never shrinks.
type CompactScratch struct {
	flats       []int
	chunkCounts []int
	holeBits    []uint64
	moves       []SlotMove
}

type compactMode int

const (
	modeAllMigrate compactMode = iota // n == Len(): zero-and-free all, no SlotMoves
	modeChunkSwap                     // ≥1 full-hole chunk: chunk swap + optional slot pass
	modeSlotLevel                     // only partial holes: sort + two-pointer
)

// Defragmenter fills the holes left by bulk entity migration, choosing the
// cheapest compaction strategy per call. Embed by value in long-lived callers.
type Defragmenter struct {
	scratch CompactScratch
}

// Compact fills holes in t and returns the resulting relocations. holes need
// not be sorted and may be reordered in place; the returned slice is scratch-
// backed and must be consumed before the next Compact call.
func (d *Defragmenter) Compact(t *Table, holes []SlotRef) []SlotMove {
	if len(holes) == 0 {
		return nil
	}
	t.version++
	chunkCap := int(t.chunkPack.Layout.ChunkCap)
	d.scratch.moves = d.scratch.moves[:0]

	switch d.classify(t, holes, chunkCap) {
	case modeAllMigrate:
		d.compactAllMigrate(t)
		return nil
	case modeChunkSwap:
		d.compactChunkThenSlot(t, holes, chunkCap)
	case modeSlotLevel:
		d.compactSlots(t, holes, chunkCap)
	}
	return d.scratch.moves
}

// classify picks the compaction mode; fills scratch.chunkCounts as a side effect.
func (d *Defragmenter) classify(t *Table, holes []SlotRef, chunkCap int) compactMode {
	n := len(holes)
	if uint32(n) == t.chunkPack.Len() {
		return modeAllMigrate
	}

	maxIdx := Idx(0)
	for _, h := range holes {
		if h.Idx > maxIdx {
			maxIdx = h.Idx
		}
	}
	numChunks := int(maxIdx) + 1

	if cap(d.scratch.chunkCounts) < numChunks {
		d.scratch.chunkCounts = make([]int, numChunks)
	} else {
		d.scratch.chunkCounts = d.scratch.chunkCounts[:numChunks]
		clear(d.scratch.chunkCounts)
	}

	hasFullChunk := false
	for _, h := range holes {
		idx := int(h.Idx)
		d.scratch.chunkCounts[idx]++
		if d.scratch.chunkCounts[idx] == chunkCap {
			hasFullChunk = true
		}
	}
	if hasFullChunk {
		return modeChunkSwap
	}
	return modeSlotLevel
}

// compactAllMigrate zeroes and frees whole chunks — every entity is leaving,
// nothing relocates.
func (d *Defragmenter) compactAllMigrate(t *Table) {
	for idx := Idx(0); idx < t.chunkPack.NumChunks(); idx++ {
		n := int(t.chunkPack.ChunkLen(idx))
		if n == 0 {
			continue
		}
		t.zeroRange(t.chunkPack.ChunkPtr(idx), 0, n)
		t.chunkPack.BulkFreeChunk(idx)
	}
}

// compactChunkThenSlot handles ≥1 fully vacated chunk: phase A swaps live
// chunks into full-hole positions via SwapChunks — no byte copies, and no
// address-book updates since the book stores chunk pointers, not indices —
// then phase B slot-compacts the residual partial holes.
func (d *Defragmenter) compactChunkThenSlot(t *Table, holes []SlotRef, chunkCap int) {
	numChunks := len(d.scratch.chunkCounts)

	// Partition in place: full-chunk holes are freed by phase A, the rest go to phase B.
	write := 0
	for _, h := range holes {
		if d.scratch.chunkCounts[int(h.Idx)] != chunkCap {
			holes[write] = h
			write++
		}
	}
	partialHoles := holes[:write]

	// Freed chunks are block-zeroed first: the memory is reused by later
	// allocations, and freed slots must uphold the "freed slot = zeroed" invariant.
	zeroFree := func(idx Idx) {
		t.zeroRange(t.chunkPack.ChunkPtr(idx), 0, chunkCap)
		t.chunkPack.BulkFreeChunk(idx)
	}

	leftIdx := 0
	rightIdx := numChunks - 1

	for leftIdx < rightIdx {
		for rightIdx > leftIdx && d.scratch.chunkCounts[rightIdx] == chunkCap {
			zeroFree(Idx(rightIdx))
			rightIdx--
		}
		if leftIdx >= rightIdx {
			break
		}

		for leftIdx < rightIdx && d.scratch.chunkCounts[leftIdx] != chunkCap {
			leftIdx++
		}
		if leftIdx >= rightIdx {
			break
		}

		// leftIdx is full-hole, rightIdx is live: swap metadata, no byte copies.
		t.chunkPack.SwapChunks(Idx(leftIdx), Idx(rightIdx))
		zeroFree(Idx(rightIdx))
		d.scratch.chunkCounts[leftIdx] = d.scratch.chunkCounts[rightIdx]
		for i := range partialHoles {
			if partialHoles[i].Idx == Idx(rightIdx) {
				partialHoles[i].Idx = Idx(leftIdx)
			}
		}

		leftIdx++
		rightIdx--
	}

	// Pointers may converge on a full-hole chunk with no live partner left.
	if leftIdx == rightIdx && leftIdx < numChunks &&
		d.scratch.chunkCounts[leftIdx] == chunkCap {
		zeroFree(Idx(leftIdx))
	}

	d.compactSlots(t, partialHoles, chunkCap)
}

// compactSlots fills partial holes: contiguous block fast path first, else
// sorted two-pointer (tail entity moves into each ascending hole). Zeroing of
// the vacated tail is deferred and done in blocks — the "freed slot = zeroed"
// invariant is violated only inside this function and restored on return.
func (d *Defragmenter) compactSlots(t *Table, holes []SlotRef, chunkCap int) {
	n := len(holes)
	if n == 0 {
		return
	}

	if d.compactSlotsContiguous(t, holes) {
		return
	}

	if cap(d.scratch.flats) < n {
		d.scratch.flats = make([]int, n)
	}
	hFlats := d.scratch.flats[:n]

	for i, h := range holes {
		hFlats[i] = int(h.Idx)*chunkCap + int(h.Slot)
	}
	if !sort.IntsAreSorted(hFlats) {
		sort.Ints(hFlats)
	}

	// Hole-membership bitmap: O(1) isHole on every tail retreat.
	maxFlat := hFlats[n-1]
	words := maxFlat>>6 + 1
	if cap(d.scratch.holeBits) < words {
		d.scratch.holeBits = make([]uint64, words)
	} else {
		d.scratch.holeBits = d.scratch.holeBits[:words]
		clear(d.scratch.holeBits)
	}
	for _, f := range hFlats {
		d.scratch.holeBits[f>>6] |= 1 << (uint(f) & 63)
	}
	isHole := func(f int) bool {
		return f <= maxFlat && d.scratch.holeBits[f>>6]&(1<<(uint(f)&63)) != 0
	}

	tci, ts := t.chunkPack.ResolveTail()
	ti := int(tci)
	tSlot := int(ts)
	tailF := ti*chunkCap + tSlot
	tPtr := t.chunkPack.ChunkPtr(Idx(ti))
	enterLen := tSlot + 1 // occupied length of the current tail chunk on entry

	freeHead := func() {
		t.chunkPack.FreeSlot(Idx(ti))
		if tSlot > 0 {
			tSlot--
		} else if ti > 0 {
			// Tail leaves a fully drained chunk — block-zero it now.
			t.zeroRange(tPtr, 0, enterLen)
			ti--
			tSlot = int(t.chunkPack.ChunkLen(Idx(ti))) - 1
			tPtr = t.chunkPack.ChunkPtr(Idx(ti))
			enterLen = tSlot + 1
		}
		tailF = ti*chunkCap + tSlot
	}

	holeChunk := -1
	var holePtr unsafe.Pointer
	for _, hf := range hFlats {
		for tailF > hf && isHole(tailF) {
			freeHead()
		}
		if tailF <= hf {
			if tailF == hf {
				freeHead()
			}
			break
		}
		if hc := hf / chunkCap; hc != holeChunk {
			holeChunk = hc
			holePtr = t.chunkPack.ChunkPtr(Idx(hc))
		}
		holeSlot := Slot(hf % chunkCap)
		movedID := *(*uid.UID64)(t.columns[entityColumnPos].At(tPtr, Slot(tSlot)))
		t.swapCopy(holePtr, holeSlot, tPtr, Slot(tSlot))
		freeHead()
		d.scratch.moves = append(d.scratch.moves, SlotMove{
			ID:      movedID,
			NewPtr:  holePtr,
			NewSlot: holeSlot,
		})
	}

	// Zero the vacated suffix of the final tail chunk.
	if newLen := int(t.chunkPack.ChunkLen(Idx(ti))); enterLen > newLen {
		t.zeroRange(tPtr, Slot(newLen), enterLen-newLen)
	}
}

// compactSlotsContiguous handles the dominant shape — all holes form one
// ascending contiguous range in the tail chunk — with one block shift and one
// block zero, no per-slot work. Returns false (t untouched) otherwise.
func (d *Defragmenter) compactSlotsContiguous(t *Table, holes []SlotRef) bool {
	first := holes[0]
	n := len(holes)
	for k := 1; k < n; k++ {
		if holes[k].Idx != first.Idx || holes[k].Slot != first.Slot+Slot(k) {
			return false
		}
	}
	// The hole chunk must be the pack's tail chunk.
	tailIdx, _ := t.chunkPack.ResolveTail()
	if tailIdx != first.Idx {
		return false
	}

	chunkLen := int(t.chunkPack.ChunkLen(first.Idx))
	holeEnd := int(first.Slot) + n
	moveN := chunkLen - holeEnd

	// Cost gate: the shift relocates moveN survivors, the fallback relocates
	// one entity per hole — shift only when it wins on address-book traffic.
	if moveN > n {
		return false
	}

	if moveN > 0 {
		t.moveRange(first.Ptr, first.Slot, Slot(holeEnd), moveN)
		movedIDs := unsafe.Slice(
			(*uid.UID64)(t.columns[entityColumnPos].At(first.Ptr, first.Slot)), moveN)
		for k, id := range movedIDs {
			d.scratch.moves = append(d.scratch.moves, SlotMove{
				ID:      id,
				NewPtr:  first.Ptr,
				NewSlot: first.Slot + Slot(k),
			})
		}
	}
	t.zeroRange(first.Ptr, Slot(int(first.Slot)+moveN), n)
	t.chunkPack.FreeSlots(first.Idx, n)
	return true
}
