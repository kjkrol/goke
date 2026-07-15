package colstore

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/comp"
)

// --- Defragmenter.Compact ---

func TestTable_CompactHoles_Nil_NoOp(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	tbl.SpawnCursor(cur, 0, 3, baked)

	var d Defragmenter
	moves := d.Compact(tbl, nil)

	if moves != nil {
		t.Errorf("expected nil moves for nil holes, got %v", moves)
	}
	if tbl.Len() != 3 {
		t.Errorf("expected Len unchanged at 3, got %d", tbl.Len())
	}
}

func TestTable_CompactHoles_TailHoles_NoMoves(t *testing.T) {
	// Holes are the last two slots — they are freed with no entity movement.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	tbl.SpawnCursor(cur, 0, 5, baked)

	ptr0 := tbl.ChunkPtrAt(0)
	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 3}, {Ptr: ptr0, Idx: 0, Slot: 4}}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if len(moves) != 0 {
		t.Errorf("expected 0 moves for tail holes, got %d", len(moves))
	}
	if tbl.Len() != 3 {
		t.Errorf("expected Len 3, got %d", tbl.Len())
	}
}

func TestTable_CompactHoles_FrontHoles(t *testing.T) {
	// Holes at slots 0 and 1 — tail entities fill them.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	ids, _ := tbl.SpawnCursor(cur, 0, 5, baked) // IDs 1..5
	// ids is backed by the entity column; copy values before compaction modifies them.
	savedIDs := make([]uid.UID64, len(ids))
	copy(savedIDs, ids)

	ptr0 := tbl.ChunkPtrAt(0)
	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 0}, {Ptr: ptr0, Idx: 0, Slot: 1}}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 3 {
		t.Errorf("expected Len 3, got %d", tbl.Len())
	}
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(moves))
	}
	// tail entity (savedIDs[4]) → slot 0; next tail (savedIDs[3]) → slot 1
	if moves[0].ID != savedIDs[4] || moves[0].NewPtr != ptr0 || moves[0].NewSlot != 0 {
		t.Errorf("move[0]: expected savedIDs[4]→slot0, got ID=%v ptr=%v slot=%v", moves[0].ID, moves[0].NewPtr, moves[0].NewSlot)
	}
	if moves[1].ID != savedIDs[3] || moves[1].NewPtr != ptr0 || moves[1].NewSlot != 1 {
		t.Errorf("move[1]: expected savedIDs[3]→slot1, got ID=%v ptr=%v slot=%v", moves[1].ID, moves[1].NewPtr, moves[1].NewSlot)
	}
}

func TestTable_CompactHoles_MiddleHole(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	ids, _ := tbl.SpawnCursor(cur, 0, 5, baked)
	savedIDs := make([]uid.UID64, len(ids))
	copy(savedIDs, ids)

	ptr0 := tbl.ChunkPtrAt(0)
	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 2}}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 4 {
		t.Errorf("expected Len 4, got %d", tbl.Len())
	}
	if len(moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(moves))
	}
	if moves[0].ID != savedIDs[4] || moves[0].NewPtr != ptr0 || moves[0].NewSlot != 2 {
		t.Errorf("expected savedIDs[4]→slot2, got ID=%v ptr=%v slot=%v", moves[0].ID, moves[0].NewPtr, moves[0].NewSlot)
	}
}

func TestTable_CompactHoles_TailIsHole_SkippedBeforeFill(t *testing.T) {
	// Hole at slot 1 and hole at slot 4 (tail). Tail is freed first;
	// the entity at slot 3 fills slot 1.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	ids, _ := tbl.SpawnCursor(cur, 0, 5, baked)
	savedIDs := make([]uid.UID64, len(ids))
	copy(savedIDs, ids)

	ptr0 := tbl.ChunkPtrAt(0)
	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 1}, {Ptr: ptr0, Idx: 0, Slot: 4}}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 3 {
		t.Errorf("expected Len 3, got %d", tbl.Len())
	}
	if len(moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(moves))
	}
	if moves[0].ID != savedIDs[3] || moves[0].NewPtr != ptr0 || moves[0].NewSlot != 1 {
		t.Errorf("expected savedIDs[3]→slot1, got ID=%v ptr=%v slot=%v", moves[0].ID, moves[0].NewPtr, moves[0].NewSlot)
	}
}

func TestTable_Compact_ScatteredHoles_VacatedTailZeroed(t *testing.T) {
	// Scattered (non-contiguous) holes take the two-pointer path with deferred
	// block zeroing. After Compact the vacated tail slots must be fully zeroed —
	// the "freed slot = zeroed" invariant is restored before returning.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	tbl.SpawnCursor(cur, 0, 5, baked)

	ptr0 := tbl.ChunkPtrAt(0)
	// Fill the component column with non-zero data in every slot.
	for s := range Slot(5) {
		*(*uint64)(tbl.ComponentAt(ptr0, s, 1)) = 0xDEADBEEF
	}

	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 1}, {Ptr: ptr0, Idx: 0, Slot: 3}}
	var d Defragmenter
	d.Compact(tbl, holes)

	if tbl.Len() != 3 {
		t.Fatalf("expected Len 3, got %d", tbl.Len())
	}
	// Slots 3 and 4 are the vacated tail — both columns must be zero.
	for s := Slot(3); s < 5; s++ {
		if got := *(*uint64)(tbl.ComponentAt(ptr0, s, 1)); got != 0 {
			t.Errorf("slot %d: component not zeroed after compaction, got %#x", s, got)
		}
		if got := *(*uid.UID64)(tbl.columns[entityColumnPos].At(ptr0, s)); got != 0 {
			t.Errorf("slot %d: entity ID not zeroed after compaction, got %v", s, got)
		}
	}
}

func TestTable_CompactHoles_AllHoles_EmptyTable(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	tbl.SpawnCursor(cur, 0, 3, baked)

	ptr0 := tbl.ChunkPtrAt(0)
	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 0}, {Ptr: ptr0, Idx: 0, Slot: 1}, {Ptr: ptr0, Idx: 0, Slot: 2}}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if len(moves) != 0 {
		t.Errorf("expected 0 moves when all entities leave, got %d", len(moves))
	}
	if tbl.Len() != 0 {
		t.Errorf("expected empty table, got Len %d", tbl.Len())
	}
}
