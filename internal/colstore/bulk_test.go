package colstore

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/comp"
)

// --- AllocSlots ---

func TestTable_AllocSlots_DoesNotSeedIDs(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	var tbl Table
	tbl.Init(defs)
	seederCalled := false
	tbl.SetIDSeeder(func(_ []uid.UID64, _ unsafe.Pointer, _ Slot) { seederCalled = true })

	firstIdx, _, _ := tbl.ReserveSlots(3)
	ptr, startSlot := tbl.AllocSlots(firstIdx, 3)
	tbl.ReleaseSlots()

	if seederCalled {
		t.Error("AllocSlots must not invoke the IDSeeder")
	}
	if ptr == nil {
		t.Error("expected non-nil base pointer")
	}
	if startSlot != 0 {
		t.Errorf("expected startSlot 0, got %d", startSlot)
	}
	if tbl.Len() != 3 {
		t.Errorf("expected Len 3 after allocating 3 slots, got %d", tbl.Len())
	}
}

// --- SetEntityRange ---

func TestTable_SetEntityRange_WrittenIDsVisibleViaCursor(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)

	firstIdx, _, _ := tbl.ReserveSlots(2)
	ptr, slot := tbl.AllocSlots(firstIdx, 2)
	tbl.ReleaseSlots()

	want := []uid.UID64{42, 43}
	tbl.SetEntityRange(ptr, slot, want)

	cur := newCursor(1)
	offsets := tbl.BakeOffsets([]comp.ID{1})
	_, ok := tbl.FillCursorNext(cur, 0, offsets)
	if !ok {
		t.Fatal("expected a non-empty chunk after AllocSlots+SetEntityRange")
	}
	if len(cur.IDs) != 2 || cur.IDs[0] != want[0] || cur.IDs[1] != want[1] {
		t.Errorf("expected entity IDs %v via cursor, got %v", want, cur.IDs)
	}
}

// --- CopyRangeFrom ---

type posData struct{ X, Y float32 }
type accData struct{ Z float32 }

func TestTable_CopyRangeFrom_CopiesSharedColumns(t *testing.T) {
	// src has comp 1 (posData) + comp 2 (velData).
	// dst has comp 1 (posData) + comp 3 (accData).
	// After CopyRangeFrom of 2 slots: dst.comp1 == src.comp1 for both slots;
	// dst.comp3 stays zero.
	posID := comp.ID(1)
	velID := comp.ID(2)
	accID := comp.ID(3)

	srcDefs := []comp.Def{{ID: posID, Size: 8, Align: 4}, {ID: velID, Size: 4, Align: 4}}
	dstDefs := []comp.Def{{ID: posID, Size: 8, Align: 4}, {ID: accID, Size: 4, Align: 4}}

	src := newTestTable(t, srcDefs)
	dst := newTestTable(t, dstDefs)

	srcBaked := src.BakeColumns(srcDefs)
	srcCur := newCursor(2)
	_, srcPos := src.SpawnCursor(srcCur, 0, 2, srcBaked)
	srcPtr := src.ChunkPtrAt(srcPos.Idx)
	*(*posData)(src.ComponentAt(srcPtr, srcPos.Slot, posID)) = posData{X: 1.5, Y: 2.5}
	*(*posData)(src.ComponentAt(srcPtr, srcPos.Slot+1, posID)) = posData{X: 3.5, Y: 4.5}

	dstFirstIdx, _, _ := dst.ReserveSlots(2)
	dstPtr, dstSlot := dst.AllocSlots(dstFirstIdx, 2)
	dst.ReleaseSlots()

	dst.CopyRangeFrom(src, srcPtr, srcPos.Slot, dstPtr, dstSlot, 2)

	if got := *(*posData)(dst.ComponentAt(dstPtr, dstSlot, posID)); got != (posData{X: 1.5, Y: 2.5}) {
		t.Errorf("slot 0: expected posData{1.5,2.5} in dst, got %+v", got)
	}
	if got := *(*posData)(dst.ComponentAt(dstPtr, dstSlot+1, posID)); got != (posData{X: 3.5, Y: 4.5}) {
		t.Errorf("slot 1: expected posData{3.5,4.5} in dst, got %+v", got)
	}
	gotAcc := *(*accData)(dst.ComponentAt(dstPtr, dstSlot, accID))
	if gotAcc != (accData{}) {
		t.Errorf("expected zero accData in dst (not in src), got %+v", gotAcc)
	}
}

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
