package chunk

import "testing"

// ResolveTail: a freshly initialized Pack has exactly one (empty) chunk and
// nothing to trim.
func TestResolveTail_FreshPackKeepsSingleChunk(t *testing.T) {
	g := newTestPack(t)

	g.ResolveTail()

	if len(g.chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(g.chunks))
	}
}

// ResolveTail: draining a table that spilled into extra chunks trims them
// back down to the single floor chunk at index 0 (it is never trimmed away).
func TestResolveTail_FullDrainTrimsBackToFloorChunk(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.Extend(0, cap) // fill chunk 0
	g.AllocSlot()    // spills into a new chunk 1

	if len(g.chunks) != 2 {
		t.Fatalf("expected 2 chunks after spill, got %d", len(g.chunks))
	}

	// Drain mirrors RemoveAt: resolve the tail before freeing each slot.
	g.ResolveTail()
	g.FreeSlot(1) // empty chunk 1 — trimmed by the next ResolveTail call
	for i := 0; i < cap; i++ {
		g.ResolveTail()
		g.FreeSlot(0)
	}

	g.ResolveTail()

	if len(g.chunks) != 1 {
		t.Errorf("expected drain to trim back to the single floor chunk, got %d", len(g.chunks))
	}
}

// ResolveTail never trims at or below the Reserved floor.
func TestResolveTail_RespectsReservedFloor(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.ReserveSlots(cap * 2) // reserves chunks 0 and 1, Reserved == 1

	g.ResolveTail()

	if len(g.chunks) < 2 {
		t.Errorf("expected reserved chunks to survive, got %d chunks", len(g.chunks))
	}
}

// Purge trims trailing empty chunks on demand, without needing a RemoveAt
// (and its implicit ResolveTail call) to trigger it.
func TestPurge_TrimsEmptyTrailingChunksWithoutRemoveAt(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.Extend(0, cap) // fill chunk 0
	g.AllocSlot()    // spills into chunk 1
	g.FreeSlot(1)    // empty chunk 1 directly — no ResolveTail involved

	if len(g.chunks) != 2 {
		t.Fatalf("expected chunk 1 to still be allocated before Purge, got %d chunks", len(g.chunks))
	}

	g.Purge()

	if len(g.chunks) != 1 {
		t.Errorf("expected Purge to release the empty trailing chunk, got %d", len(g.chunks))
	}
}

// Purge never trims at or below the Reserved floor, same as ResolveTail.
func TestPurge_RespectsReservedFloor(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.ReserveSlots(cap * 2)
	before := len(g.chunks)

	g.Purge()

	if len(g.chunks) != before {
		t.Errorf("expected Purge to leave reserved chunks alone, got %d (was %d)", len(g.chunks), before)
	}
}
