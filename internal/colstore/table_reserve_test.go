package colstore

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

// ReserveSlots/ReleaseSlots: thin delegates to chunk.Pack. The detailed
// chunk-growth/trim semantics are covered in internal/chunk; here we only
// verify Table wires the call through correctly.
func TestTable_ReserveAndReleaseSlots(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)

	firstIdx, firstAvailable, chunkCap := tbl.ReserveSlots(5)
	if firstIdx != 0 {
		t.Errorf("expected firstIdx 0 on a fresh table, got %d", firstIdx)
	}
	if firstAvailable != chunkCap {
		t.Errorf("expected firstAvailable == chunkCap (%d) on a fresh table, got %d", chunkCap, firstAvailable)
	}

	// Reserving more than one chunk's worth must spill into extra chunks and
	// move the Reserved floor off its zero-value default.
	tbl2 := newTestTable(t, defs)
	tbl2.ReserveSlots(chunkCap*2 + 1)
	if tbl2.chunkPack.Reserved == 0 {
		t.Error("expected ReserveSlots to move the Reserved floor when spilling into more chunks")
	}

	tbl2.ReleaseSlots()
	if tbl2.chunkPack.Reserved != 0 {
		t.Errorf("expected Reserved cleared after ReleaseSlots, got %d", tbl2.chunkPack.Reserved)
	}
}

// Purge: thin delegate to chunk.Pack.Purge. The chunk-trimming mechanics
// themselves are covered in internal/chunk; Table exposes no way to observe
// chunk count, so here we only confirm the call reaches chunkPack and leaves
// Table's own state (Len) intact.
func TestTable_Purge(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)

	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	_, pos := tbl.SpawnCursor(cur, 0, 1, baked)
	tbl.RemoveAt(pos)

	tbl.Purge()

	if tbl.Len() != 0 {
		t.Errorf("expected Len to remain 0 after Purge, got %d", tbl.Len())
	}
}
