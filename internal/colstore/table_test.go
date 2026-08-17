package colstore

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
)

// newTestTable builds a Table for compDefs with a seeder that hands out
// sequential, distinct entity IDs starting at 1.
func newTestTable(t *testing.T, compDefs []comp.Def) *Table {
	t.Helper()
	var tbl Table
	tbl.Init(compDefs)
	next := uint64(1)
	tbl.SetIDSeeder(func(dst []uid.UID64, _ unsafe.Pointer, _ Slot) {
		for i := range dst {
			dst[i] = uid.UID64(next)
			next++
		}
	})
	return &tbl
}

// newCursor returns an iter.Cursor with Offsets sized for n tracked columns.
func newCursor(n int) *iter.Cursor {
	return &iter.Cursor{Offsets: make([]uintptr, n)}
}

func TestTable_LenTracking(t *testing.T) {
	compDefs := []comp.Def{
		{ID: 1, Size: 8, Align: 8},
	}

	var cs Table
	cs.Init(compDefs)

	if cs.Len() != 0 {
		t.Errorf("Expected initial Table.Len to be 0, got %d", cs.Len())
	}

	cs.chunkPack.AllocSlot()
	cs.chunkPack.AllocSlot()
	cs.chunkPack.AllocSlot()

	if cs.Len() != 3 {
		t.Errorf("Expected Table.Len to be 3 after 3 allocations, got %d", cs.Len())
	}

	if cs.chunkPack.ChunkLen(0) != 3 {
		t.Errorf("Expected chunk.Len to be 3, got %d", cs.chunkPack.ChunkLen(0))
	}

	cs.Clear()
	if cs.Len() != 0 {
		t.Errorf("Expected Table.Len to be 0 after Clear, got %d", cs.Len())
	}
}

func TestTable_Version_BumpsOnCompact(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	tbl.SpawnCursor(cur, 0, 3, baked)

	before := tbl.Version()

	ptr0 := tbl.ChunkPtrAt(0)
	var d Defragmenter
	d.Compact(tbl, []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 0}})

	if tbl.Version() != before+1 {
		t.Errorf("expected Version to bump by 1 after Compact, got %d (was %d)", tbl.Version(), before)
	}
}

func TestTable_ChunkIdxByPtr(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	chunkCap := int(tbl.chunkPack.Layout.ChunkCap)

	cur := newCursor(1)
	tbl.SpawnCursor(cur, 0, chunkCap, baked)
	tbl.chunkPack.AddChunks(1)
	cur2 := newCursor(1)
	tbl.SpawnCursor(cur2, 1, 2, baked)

	ptr1 := tbl.ChunkPtrAt(1)
	if got := tbl.ChunkIdxByPtr(ptr1); got != 1 {
		t.Errorf("expected ChunkIdxByPtr to return 1, got %d", got)
	}
}

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
