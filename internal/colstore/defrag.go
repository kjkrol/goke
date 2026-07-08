package colstore

import (
	"sort"
	"unsafe"

	"github.com/kjkrol/uid"
)

// CompactScratch holds all working buffers used by Defragmenter.
// Slices grow lazily and never shrink; zero allocations after warm-up.
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
// cheapest compaction strategy via a single decision switch.
// Embed by value inside long-lived callers (e.g. Migrator) to reuse scratch
// buffers across calls.
type Defragmenter struct {
	scratch CompactScratch
}

// Compact fills holes in t and returns SlotMoves for addr.Book updates.
// holes contains source positions vacated by entities already copied to dst;
// it need not be sorted. The returned slice is backed by internal scratch and
// must be consumed before the next Compact call. holes may be reordered in place.
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

// classify determines the compaction strategy.
// modeAllMigrate is checked first (O(1) comparison).
// Then holes are scanned twice using the pre-filled h.Idx to find the
// highest-indexed chunk and count holes per chunk; modeChunkSwap is returned
// if any chunk is entirely holes.
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

// compactAllMigrate handles the case where every entity in t is leaving.
// Each occupied chunk is zeroed with one ZeroMemory per column and freed
// whole; no entity needs to relocate.
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

// compactChunkThenSlot handles the case where at least one entire chunk is vacated.
//
// Phase A (chunk-level two-pointer): each full-hole chunk is paired with the
// rightmost non-full-hole chunk. SwapChunks relocates the live chunk without
// copying any component bytes. Because addr.Book stores chunk memory pointers
// (not indices), the swapped entities' book entries remain valid with no updates.
// After each swap, h.Idx in partialHoles is updated to reflect the new position.
//
// Phase B (slot-level): residual partial holes are filled using the standard
// sort + two-pointer algorithm.
func (d *Defragmenter) compactChunkThenSlot(t *Table, holes []SlotRef, chunkCap int) {
	numChunks := len(d.scratch.chunkCounts)

	// Partition holes[] in-place: drop full-chunk entries (freed in Phase A),
	// keep partial-chunk entries for Phase B.
	write := 0
	for _, h := range holes {
		if d.scratch.chunkCounts[int(h.Idx)] != chunkCap {
			holes[write] = h
			write++
		}
	}
	partialHoles := holes[:write]

	// Phase A: chunk-level two-pointer.
	// No entity ID reads. No addr.Book updates — ptrs remain valid after SwapChunks.
	// Every freed full-hole chunk is block-zeroed first, upholding the invariant
	// that freed slots are zeroed (chunk memory is reused by later allocations,
	// and columns absent in a migration source are never overwritten).
	zeroFree := func(idx Idx) {
		t.zeroRange(t.chunkPack.ChunkPtr(idx), 0, chunkCap)
		t.chunkPack.BulkFreeChunk(idx)
	}

	leftIdx := 0
	rightIdx := numChunks - 1

	for leftIdx < rightIdx {
		// Retreat right past trailing full-hole chunks — free them directly.
		for rightIdx > leftIdx && d.scratch.chunkCounts[rightIdx] == chunkCap {
			zeroFree(Idx(rightIdx))
			rightIdx--
		}
		if leftIdx >= rightIdx {
			break
		}

		// Advance left to the next full-hole chunk.
		for leftIdx < rightIdx && d.scratch.chunkCounts[leftIdx] != chunkCap {
			leftIdx++
		}
		if leftIdx >= rightIdx {
			break
		}

		// leftIdx is full-hole; rightIdx is non-full-hole (live).
		// Swap chunk metadata: leftIdx receives rightIdx's data, no byte copies.
		t.chunkPack.SwapChunks(Idx(leftIdx), Idx(rightIdx))

		// Free the now-vacated right position and propagate hole count.
		zeroFree(Idx(rightIdx))
		d.scratch.chunkCounts[leftIdx] = d.scratch.chunkCounts[rightIdx]

		// Update Idx for partial holes that were in the swapped-in chunk.
		for i := range partialHoles {
			if partialHoles[i].Idx == Idx(rightIdx) {
				partialHoles[i].Idx = Idx(leftIdx)
			}
		}

		leftIdx++
		rightIdx--
	}

	// If both pointers converged on the same full-hole chunk with no live partner, free it.
	if leftIdx == rightIdx && leftIdx < numChunks &&
		d.scratch.chunkCounts[leftIdx] == chunkCap {
		zeroFree(Idx(leftIdx))
	}

	// Phase B: slot-level compaction for residual partial holes.
	d.compactSlots(t, partialHoles, chunkCap)
}

// compactSlots fills partial holes, preferring the contiguous block fast path
// and falling back to the sorted two-pointer algorithm: for each hole
// (ascending flat index) the tail entity moves into the hole and the vacated
// tail slot is freed.
//
// Zeroing of vacated slots is deferred: the tail retreats monotonically, so
// the vacated region is a contiguous suffix — fully drained chunks are
// block-zeroed when the tail crosses out of them, and the final chunk's
// vacated range once after the loop. Freed slots hold stale bytes only inside
// this function; the "freed slot = zeroed" invariant is restored on return.
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

	// Hole-membership bitmap indexed by flat slot — O(1) isHole instead of a
	// binary search on every tail retreat (64 slots per uint64 word).
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
			// Leaving a fully drained chunk: block-zero its former occupied range.
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

	// Block-zero the vacated suffix of the final tail chunk.
	if newLen := int(t.chunkPack.ChunkLen(Idx(ti))); enterLen > newLen {
		t.zeroRange(tPtr, Slot(newLen), enterLen-newLen)
	}
}

// compactSlotsContiguous handles the dominant bulk-migration shape: all holes
// form one ascending contiguous slot range inside the pack's tail chunk (the
// natural result of migrating a consecutive run captured from Query.All()).
// The surviving block above the holes shifts down with one CopyMemory per
// column and the vacated tail range is zeroed with one ZeroMemory per column —
// no per-slot work. Returns false (leaving t untouched) when the holes don't
// match this shape.
func (d *Defragmenter) compactSlotsContiguous(t *Table, holes []SlotRef) bool {
	first := holes[0]
	n := len(holes)
	for k := 1; k < n; k++ {
		if holes[k].Idx != first.Idx || holes[k].Slot != first.Slot+Slot(k) {
			return false
		}
	}
	// Survivors above the hole range must live in the same chunk, i.e. the
	// hole chunk is the pack's tail chunk.
	tailIdx, _ := t.chunkPack.ResolveTail()
	if tailIdx != first.Idx {
		return false
	}

	chunkLen := int(t.chunkPack.ChunkLen(first.Idx))
	holeEnd := int(first.Slot) + n
	moveN := chunkLen - holeEnd

	// The block shift relocates every survivor above the holes (moveN book
	// updates); the two-pointer fallback relocates one entity per hole (n book
	// updates). Shift only when it doesn't lose on addr.Book traffic — for a
	// small hole range deep in the chunk the fallback is cheaper.
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
