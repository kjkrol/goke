package colstore

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/comp"
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

func TestTable_CompactHoles_FullChunkSwappedWithSurvivorChunk(t *testing.T) {
	// Chunk 0 fully drains (every slot a hole) while chunk 1 keeps some
	// survivors alongside a couple of its own holes — exercises
	// compactChunkThenSlot's full-hole-chunk swap phase together with the
	// compactSlots pass it runs afterward on the remapped partial holes.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	chunkCap := int(tbl.chunkPack.Layout.ChunkCap)

	cur0 := newCursor(1)
	tbl.SpawnCursor(cur0, 0, chunkCap, baked)

	tbl.chunkPack.AddChunks(1)
	cur1 := newCursor(1)
	ids1, _ := tbl.SpawnCursor(cur1, 1, 5, baked)
	savedIDs1 := make([]uid.UID64, len(ids1))
	copy(savedIDs1, ids1)

	ptr1 := tbl.ChunkPtrAt(1)
	for s := range Slot(5) {
		*(*uint64)(tbl.ComponentAt(ptr1, s, 1)) = uint64(100 + s)
	}

	ptr0 := tbl.ChunkPtrAt(0)
	holes := make([]SlotRef, 0, chunkCap+2)
	for s := 0; s < chunkCap; s++ {
		holes = append(holes, SlotRef{Ptr: ptr0, Idx: 0, Slot: Slot(s)})
	}
	holes = append(holes, SlotRef{Ptr: ptr1, Idx: 1, Slot: 1}, SlotRef{Ptr: ptr1, Idx: 1, Slot: 3})

	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 3 {
		t.Fatalf("expected Len 3 after compaction, got %d", tbl.Len())
	}
	if got := tbl.chunkPack.NumChunks(); got != 1 {
		t.Errorf("expected the fully-drained chunk to be trimmed away, NumChunks = %d", got)
	}
	if len(moves) != 1 {
		t.Errorf("expected 1 SlotMove (tail entity filling the one non-tail hole), got %d", len(moves))
	}

	survivingPtr := tbl.ChunkPtrAt(0)
	want := map[uid.UID64]uint64{
		savedIDs1[0]: 100,
		savedIDs1[2]: 102,
		savedIDs1[4]: 104,
	}
	got := make(map[uid.UID64]uint64, 3)
	for s := range Slot(3) {
		id := *(*uid.UID64)(tbl.columns[entityColumnPos].At(survivingPtr, s))
		got[id] = *(*uint64)(tbl.ComponentAt(survivingPtr, s, 1))
	}
	for id, wantVal := range want {
		gotVal, ok := got[id]
		if !ok {
			t.Errorf("survivor %v missing after compaction", id)
			continue
		}
		if gotVal != wantVal {
			t.Errorf("survivor %v: component = %d, want %d", id, gotVal, wantVal)
		}
	}
}

func TestTable_CompactHoles_CrossChunkRetreat(t *testing.T) {
	// Chunk 1 (tail) fully drains via two holes; the tail retreat crosses
	// into chunk 0 (which has its own, separate hole) to keep filling —
	// exercises freeHead's chunk-boundary branch.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur0 := newCursor(1)
	ids0, _ := tbl.SpawnCursor(cur0, 0, 3, baked)
	saved0 := make([]uid.UID64, len(ids0))
	copy(saved0, ids0)

	tbl.chunkPack.AddChunks(1)
	cur1 := newCursor(1)
	tbl.SpawnCursor(cur1, 1, 2, baked)

	ptr0 := tbl.ChunkPtrAt(0)
	ptr1 := tbl.ChunkPtrAt(1)
	for s := range Slot(3) {
		*(*uint64)(tbl.ComponentAt(ptr0, s, 1)) = uint64(200 + s)
	}

	holes := []SlotRef{
		{Ptr: ptr1, Idx: 1, Slot: 0},
		{Ptr: ptr1, Idx: 1, Slot: 1},
		{Ptr: ptr0, Idx: 0, Slot: 1},
	}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", tbl.Len())
	}
	if len(moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(moves))
	}
	if moves[0].ID != saved0[2] || moves[0].NewPtr != ptr0 || moves[0].NewSlot != 1 {
		t.Errorf("expected saved0[2]→(chunk0,slot1), got ID=%v ptr=%v slot=%v", moves[0].ID, moves[0].NewPtr, moves[0].NewSlot)
	}
	if got := *(*uint64)(tbl.ComponentAt(ptr0, 0, 1)); got != 200 {
		t.Errorf("slot0: component = %d, want 200 (untouched)", got)
	}
	if got := *(*uint64)(tbl.ComponentAt(ptr0, 1, 1)); got != 202 {
		t.Errorf("slot1: component = %d, want 202 (moved from slot2)", got)
	}
}

func TestTable_CompactSlotsContiguous_NonTailChunk_FallsBackToGeneralPath(t *testing.T) {
	// Contiguous holes sit in chunk 0, but chunk 1 (with live survivors) is
	// the pack's tail — compactSlotsContiguous must decline (tailIdx
	// mismatch) and let the general two-pointer path handle it instead.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur0 := newCursor(1)
	tbl.SpawnCursor(cur0, 0, 3, baked)

	tbl.chunkPack.AddChunks(1)
	cur1 := newCursor(1)
	ids1, _ := tbl.SpawnCursor(cur1, 1, 2, baked)
	saved1 := make([]uid.UID64, len(ids1))
	copy(saved1, ids1)

	ptr0 := tbl.ChunkPtrAt(0)
	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 0}, {Ptr: ptr0, Idx: 0, Slot: 1}}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 3 {
		t.Fatalf("expected Len 3, got %d", tbl.Len())
	}
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(moves))
	}
	_ = saved1
}

func TestTable_CompactSlotsContiguous_MovesSurvivorsPastHoles(t *testing.T) {
	// 3 contiguous holes with only 1 survivor after them — the shift wins
	// the cost gate (moveN <= n), so the fast contiguous path actually
	// moves a survivor instead of just freeing tail slots.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur := newCursor(1)
	ids, _ := tbl.SpawnCursor(cur, 0, 6, baked)
	saved := make([]uid.UID64, len(ids))
	copy(saved, ids)

	ptr0 := tbl.ChunkPtrAt(0)
	for s := range Slot(6) {
		*(*uint64)(tbl.ComponentAt(ptr0, s, 1)) = uint64(300 + s)
	}

	holes := []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 2}, {Ptr: ptr0, Idx: 0, Slot: 3}, {Ptr: ptr0, Idx: 0, Slot: 4}}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 3 {
		t.Fatalf("expected Len 3, got %d", tbl.Len())
	}
	if len(moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(moves))
	}
	if moves[0].ID != saved[5] || moves[0].NewPtr != ptr0 || moves[0].NewSlot != 2 {
		t.Errorf("expected saved[5]→slot2, got ID=%v ptr=%v slot=%v", moves[0].ID, moves[0].NewPtr, moves[0].NewSlot)
	}
	if got := *(*uint64)(tbl.ComponentAt(ptr0, 2, 1)); got != 305 {
		t.Errorf("slot2: component = %d, want 305 (moved from slot5)", got)
	}
}

func TestTable_CompactHoles_MultipleFullHoleChunksOnTheRight(t *testing.T) {
	// Chunks 1, 2, 3 are all fully holes; chunk 0 survives untouched. The
	// right-side pre-scan frees all three directly (no swap partner needed)
	// and converges with leftIdx immediately.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	chunkCap := int(tbl.chunkPack.Layout.ChunkCap)

	cur0 := newCursor(1)
	ids0, _ := tbl.SpawnCursor(cur0, 0, 2, baked)
	saved0 := make([]uid.UID64, len(ids0))
	copy(saved0, ids0)

	tbl.chunkPack.AddChunks(3)
	cur := newCursor(1)
	tbl.SpawnCursor(cur, 1, chunkCap, baked)
	tbl.SpawnCursor(cur, 2, chunkCap, baked)
	tbl.SpawnCursor(cur, 3, chunkCap, baked)

	holes := make([]SlotRef, 0, chunkCap*3)
	for _, idx := range []Idx{1, 2, 3} {
		ptr := tbl.ChunkPtrAt(idx)
		for s := 0; s < chunkCap; s++ {
			holes = append(holes, SlotRef{Ptr: ptr, Idx: idx, Slot: Slot(s)})
		}
	}

	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if moves != nil {
		t.Errorf("expected nil moves (chunk 0 untouched, no partial holes), got %v", moves)
	}
	if tbl.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", tbl.Len())
	}
	ptr0 := tbl.ChunkPtrAt(0)
	for s := range Slot(2) {
		id := *(*uid.UID64)(tbl.columns[entityColumnPos].At(ptr0, s))
		if id != saved0[s] {
			t.Errorf("chunk0 slot %d: ID = %v, want %v (untouched)", s, id, saved0[s])
		}
	}
}

func TestTable_CompactHoles_LoneFullHoleChunkConvergesInMiddle(t *testing.T) {
	// Chunks 0 and 3 survive; chunks 1 and 2 are both fully holes. The
	// left scan must step past chunk 0 (leftIdx++) before finding chunk 1
	// to swap with chunk 3; left/right then meet exactly at chunk 2, which
	// is still a lone full-hole chunk with no remaining partner — the final
	// convergence check must free it directly.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	chunkCap := int(tbl.chunkPack.Layout.ChunkCap)

	cur0 := newCursor(1)
	ids0, _ := tbl.SpawnCursor(cur0, 0, 2, baked)
	saved0 := make([]uid.UID64, len(ids0))
	copy(saved0, ids0)

	tbl.chunkPack.AddChunks(3)
	fillCur := newCursor(1)
	tbl.SpawnCursor(fillCur, 1, chunkCap, baked)
	tbl.SpawnCursor(fillCur, 2, chunkCap, baked)
	cur3 := newCursor(1)
	ids3, _ := tbl.SpawnCursor(cur3, 3, 3, baked)
	saved3 := make([]uid.UID64, len(ids3))
	copy(saved3, ids3)

	holes := make([]SlotRef, 0, chunkCap*2)
	for _, idx := range []Idx{1, 2} {
		ptr := tbl.ChunkPtrAt(idx)
		for s := 0; s < chunkCap; s++ {
			holes = append(holes, SlotRef{Ptr: ptr, Idx: idx, Slot: Slot(s)})
		}
	}

	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if moves != nil {
		t.Errorf("expected nil moves (no partial holes anywhere), got %v", moves)
	}
	wantLen := uint32(2 + 3)
	if tbl.Len() != wantLen {
		t.Fatalf("expected Len %d, got %d", wantLen, tbl.Len())
	}

	ptr0 := tbl.ChunkPtrAt(0)
	for s := range Slot(2) {
		id := *(*uid.UID64)(tbl.columns[entityColumnPos].At(ptr0, s))
		if id != saved0[s] {
			t.Errorf("chunk0 slot %d: ID = %v, want %v (untouched)", s, id, saved0[s])
		}
	}
	ptr1 := tbl.ChunkPtrAt(1)
	got3 := make(map[uid.UID64]bool, 3)
	for s := range Slot(3) {
		got3[*(*uid.UID64)(tbl.columns[entityColumnPos].At(ptr1, s))] = true
	}
	for _, id := range saved3 {
		if !got3[id] {
			t.Errorf("chunk3 survivor %v missing from swapped-in chunk 1", id)
		}
	}
}

func TestTable_CompactHoles_LeftScanExhaustsWithoutFindingFullChunk(t *testing.T) {
	// Chunks 0 and 1 both survive untouched; chunk 2 (the tail) is fully
	// holes. The right pre-scan frees chunk 2 but stops one short of
	// leftIdx, so the left scan then walks forward (leftIdx++) and meets
	// rightIdx without ever finding a second full chunk — the second break.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	chunkCap := int(tbl.chunkPack.Layout.ChunkCap)

	cur0 := newCursor(1)
	ids0, _ := tbl.SpawnCursor(cur0, 0, 2, baked)
	saved0 := make([]uid.UID64, len(ids0))
	copy(saved0, ids0)

	tbl.chunkPack.AddChunks(2)
	cur1 := newCursor(1)
	ids1, _ := tbl.SpawnCursor(cur1, 1, 2, baked)
	saved1 := make([]uid.UID64, len(ids1))
	copy(saved1, ids1)

	cur2 := newCursor(1)
	tbl.SpawnCursor(cur2, 2, chunkCap, baked)

	ptr2 := tbl.ChunkPtrAt(2)
	holes := make([]SlotRef, 0, chunkCap)
	for s := 0; s < chunkCap; s++ {
		holes = append(holes, SlotRef{Ptr: ptr2, Idx: 2, Slot: Slot(s)})
	}

	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if moves != nil {
		t.Errorf("expected nil moves (chunks 0,1 untouched, no partial holes), got %v", moves)
	}
	if tbl.Len() != 4 {
		t.Fatalf("expected Len 4, got %d", tbl.Len())
	}
	ptr0 := tbl.ChunkPtrAt(0)
	for s := range Slot(2) {
		if id := *(*uid.UID64)(tbl.columns[entityColumnPos].At(ptr0, s)); id != saved0[s] {
			t.Errorf("chunk0 slot %d: ID = %v, want %v (untouched)", s, id, saved0[s])
		}
	}
	ptr1 := tbl.ChunkPtrAt(1)
	for s := range Slot(2) {
		if id := *(*uid.UID64)(tbl.columns[entityColumnPos].At(ptr1, s)); id != saved1[s] {
			t.Errorf("chunk1 slot %d: ID = %v, want %v (untouched)", s, id, saved1[s])
		}
	}
}

func TestTable_CompactSlots_TailRetreatSkipsPreExistingEmptyChunk(t *testing.T) {
	// Chunk 1 is already empty (never populated) sitting between chunk 0
	// and the tail chunk 2. Chunk 2 fully drains via holes, forcing the
	// tail retreat to cross a chunk boundary — freeHead's skip-loop must
	// step over the pre-existing empty chunk 1 to land correctly in chunk 0.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur0 := newCursor(1)
	ids0, _ := tbl.SpawnCursor(cur0, 0, 3, baked)
	saved0 := make([]uid.UID64, len(ids0))
	copy(saved0, ids0)

	tbl.chunkPack.AddChunks(1) // chunk 1: reserved, never populated
	tbl.chunkPack.AddChunks(1)
	cur2 := newCursor(1)
	tbl.SpawnCursor(cur2, 2, 2, baked)

	ptr0 := tbl.ChunkPtrAt(0)
	ptr2 := tbl.ChunkPtrAt(2)
	for s := range Slot(3) {
		*(*uint64)(tbl.ComponentAt(ptr0, s, 1)) = uint64(400 + s)
	}

	holes := []SlotRef{
		{Ptr: ptr0, Idx: 0, Slot: 1},
		{Ptr: ptr2, Idx: 2, Slot: 0},
		{Ptr: ptr2, Idx: 2, Slot: 1},
	}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if tbl.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", tbl.Len())
	}
	if len(moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(moves))
	}
	if moves[0].ID != saved0[2] || moves[0].NewPtr != ptr0 || moves[0].NewSlot != 1 {
		t.Errorf("expected saved0[2]→(chunk0,slot1), got ID=%v ptr=%v slot=%v", moves[0].ID, moves[0].NewPtr, moves[0].NewSlot)
	}
	if got := *(*uint64)(tbl.ComponentAt(ptr0, 1, 1)); got != 402 {
		t.Errorf("slot1: component = %d, want 402 (moved from slot2)", got)
	}
}

func TestTable_CompactAllMigrate_SkipsAlreadyEmptyChunk(t *testing.T) {
	// A pre-existing empty chunk sits alongside the chunk that's fully
	// draining — compactAllMigrate must skip it (n == 0 continue) rather
	// than zero/free it again.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur0 := newCursor(1)
	tbl.SpawnCursor(cur0, 0, 3, baked)
	tbl.chunkPack.AddChunks(1) // chunk 1 stays empty

	ptr0 := tbl.ChunkPtrAt(0)
	holes := []SlotRef{
		{Ptr: ptr0, Idx: 0, Slot: 0},
		{Ptr: ptr0, Idx: 0, Slot: 1},
		{Ptr: ptr0, Idx: 0, Slot: 2},
	}
	var d Defragmenter
	moves := d.Compact(tbl, holes)

	if moves != nil {
		t.Errorf("expected nil moves (everyone leaves), got %v", moves)
	}
	if tbl.Len() != 0 {
		t.Errorf("expected Len 0, got %d", tbl.Len())
	}
}

func TestTable_Compact_ReusesScratchAcrossCalls(t *testing.T) {
	// A second Compact call on the same Defragmenter, with numChunks no
	// larger than the first, must reuse (not reallocate) scratch.chunkCounts.
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	tbl.SpawnCursor(cur, 0, 5, baked)

	ptr0 := tbl.ChunkPtrAt(0)
	var d Defragmenter
	d.Compact(tbl, []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 0}})
	if tbl.Len() != 4 {
		t.Fatalf("after first Compact: expected Len 4, got %d", tbl.Len())
	}

	moves := d.Compact(tbl, []SlotRef{{Ptr: ptr0, Idx: 0, Slot: 1}})
	if tbl.Len() != 3 {
		t.Fatalf("after second Compact: expected Len 3, got %d", tbl.Len())
	}
	_ = moves
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
