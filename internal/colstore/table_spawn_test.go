package colstore

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

func TestTable_SpawnCursor(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur := newCursor(1)
	ids, pos := tbl.SpawnCursor(cur, 0, 3, baked)

	if len(ids) != 3 {
		t.Fatalf("expected 3 seeded IDs, got %d", len(ids))
	}
	if ids[0] == ids[1] || ids[1] == ids[2] {
		t.Errorf("expected distinct seeded IDs, got %v", ids)
	}
	if pos.Idx != 0 || pos.Slot != 0 {
		t.Errorf("expected first spawn at Pos{0,0}, got %+v", pos)
	}
	if tbl.Len() != 3 {
		t.Errorf("expected Len 3 after spawning 3 entities, got %d", tbl.Len())
	}
	if len(cur.IDs) != 3 || cur.IDs[0] != ids[0] {
		t.Errorf("expected cur.IDs to mirror the seeded IDs, got %v want %v", cur.IDs, ids)
	}
}

func TestTable_FillCursorNextAndPointCursor(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur := newCursor(1)
	ids, pos := tbl.SpawnCursor(cur, 0, 3, baked)

	idx, ok := tbl.FillCursorNext(cur, 0, []uintptr{baked[0].Offset})
	if !ok {
		t.Fatal("expected to find a non-empty chunk")
	}
	if idx != 0 {
		t.Errorf("expected chunk idx 0, got %d", idx)
	}
	if len(cur.IDs) != 3 || cur.IDs[0] != ids[0] || cur.IDs[2] != ids[2] {
		t.Errorf("expected cursor IDs %v to match spawned IDs %v", cur.IDs, ids)
	}

	// PointCursor repositions Base/Slot without touching Offsets.
	cur.Offsets = nil
	tbl.PointCursor(cur, pos)
	if cur.Base == nil {
		t.Error("expected PointCursor to set a non-nil Base")
	}
	if cur.Slot != uintptr(pos.Slot) {
		t.Errorf("expected Slot %d, got %d", pos.Slot, cur.Slot)
	}
	if cur.Offsets != nil {
		t.Error("expected PointCursor to leave Offsets untouched")
	}

	// Nothing exists past the only chunk.
	if _, ok := tbl.FillCursorNext(cur, 1, nil); ok {
		t.Error("expected no chunk found scanning from index 1")
	}
}
