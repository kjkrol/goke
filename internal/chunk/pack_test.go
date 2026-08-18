package chunk

import (
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
)

func TestPack_ChunkAccessors(t *testing.T) {
	var g Pack
	layout := Layout{}
	layout.Init([]comp.Def{{ID: 1, Size: 8, Align: 8}})
	g.Init(layout)

	if len(g.chunks) != 1 {
		t.Fatalf("expected 1 chunk after Init, got %d", len(g.chunks))
	}
	if g.ChunkLen(0) != 0 {
		t.Errorf("expected ChunkLen(0) == 0, got %d", g.ChunkLen(0))
	}
	if g.ChunkPtr(0) == nil {
		t.Error("expected non-nil ChunkPtr(0)")
	}
}

func TestPack_ExtendAndFreeSlot(t *testing.T) {
	var g Pack
	layout := Layout{}
	layout.Init([]comp.Def{{ID: 1, Size: 8, Align: 8}})
	g.Init(layout)

	g.Extend(0, 3)
	if g.ChunkLen(0) != 3 {
		t.Errorf("expected ChunkLen 3 after Extend(0, 3), got %d", g.ChunkLen(0))
	}

	g.FreeSlot(0)
	if g.ChunkLen(0) != 2 {
		t.Errorf("expected ChunkLen 2 after FreeSlot, got %d", g.ChunkLen(0))
	}
}

func TestPack_Len(t *testing.T) {
	g := newTestPack(t)

	if g.Len() != 0 {
		t.Errorf("expected Len 0 after Init, got %d", g.Len())
	}

	g.AllocSlot()
	g.AllocSlot()

	if g.Len() != 2 {
		t.Errorf("expected Len 2 after 2 AllocSlot calls, got %d", g.Len())
	}
}

func TestPack_NumChunks(t *testing.T) {
	g := newTestPack(t)
	if got := g.NumChunks(); got != 1 {
		t.Errorf("expected NumChunks 1 after Init, got %d", got)
	}
	g.AddChunks(2)
	if got := g.NumChunks(); got != 3 {
		t.Errorf("expected NumChunks 3 after AddChunks(2), got %d", got)
	}
}

func TestPack_FreeSlots(t *testing.T) {
	g := newTestPack(t)
	g.Extend(0, 5)
	if got := g.Len(); got != 5 {
		t.Fatalf("expected Len 5 after Extend, got %d", got)
	}
	g.FreeSlots(0, 3)
	if got := g.ChunkLen(0); got != 2 {
		t.Errorf("expected ChunkLen 2 after FreeSlots(0,3), got %d", got)
	}
	if got := g.Len(); got != 2 {
		t.Errorf("expected Len 2 after FreeSlots(0,3), got %d", got)
	}
}

func TestPack_SwapChunks(t *testing.T) {
	g := newTestPack(t)
	g.Extend(0, 3)
	g.AddChunks(1)
	g.Extend(1, 5)

	ptr0Before := g.ChunkPtr(0)
	ptr1Before := g.ChunkPtr(1)

	g.SwapChunks(0, 1)

	if g.ChunkPtr(0) != ptr1Before || g.ChunkLen(0) != 5 {
		t.Errorf("expected chunk 0 to now hold the old chunk 1 (ptr=%v len=5), got ptr=%v len=%d",
			ptr1Before, g.ChunkPtr(0), g.ChunkLen(0))
	}
	if g.ChunkPtr(1) != ptr0Before || g.ChunkLen(1) != 3 {
		t.Errorf("expected chunk 1 to now hold the old chunk 0 (ptr=%v len=3), got ptr=%v len=%d",
			ptr0Before, g.ChunkPtr(1), g.ChunkLen(1))
	}
}

func TestPack_BulkFreeChunk(t *testing.T) {
	g := newTestPack(t)
	g.Extend(0, 4)
	if got := g.Len(); got != 4 {
		t.Fatalf("expected Len 4, got %d", got)
	}
	g.BulkFreeChunk(0)
	if got := g.ChunkLen(0); got != 0 {
		t.Errorf("expected ChunkLen 0 after BulkFreeChunk, got %d", got)
	}
	if got := g.Len(); got != 0 {
		t.Errorf("expected Len 0 after BulkFreeChunk, got %d", got)
	}
}

func TestPack_ChunkIdxByPtr(t *testing.T) {
	g := newTestPack(t)
	g.AddChunks(2)

	ptr1 := g.ChunkPtr(1)
	if got := g.ChunkIdxByPtr(ptr1); got != 1 {
		t.Errorf("expected ChunkIdxByPtr to return 1, got %d", got)
	}

	defer func() {
		if recover() == nil {
			t.Error("expected ChunkIdxByPtr to panic for an unknown pointer")
		}
	}()
	g.ChunkIdxByPtr(nil)
}

func TestPack_NextNonEmptyChunk(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.Extend(0, cap) // fill chunk 0
	g.AllocSlot()    // spills into chunk 1, 1 slot used there
	g.FreeSlot(1)    // empty chunk 1 again, directly (no trim)

	idx, ptr, length, ok := g.NextNonEmptyChunk(0)
	if !ok {
		t.Fatal("expected to find a non-empty chunk")
	}
	if idx != 0 {
		t.Errorf("expected idx 0, got %d", idx)
	}
	if length != cap {
		t.Errorf("expected length %d, got %d", cap, length)
	}
	if ptr == nil {
		t.Error("expected non-nil ptr")
	}

	if _, _, _, ok := g.NextNonEmptyChunk(1); ok {
		t.Error("expected no non-empty chunk found starting from empty chunk 1")
	}
}

func TestPack_Clear(t *testing.T) {
	g := newTestPack(t)
	g.AllocSlot()
	g.AllocSlot()

	g.Clear()

	if g.Len() != 0 {
		t.Errorf("expected Len 0 after Clear, got %d", g.Len())
	}
	if len(g.chunks) != 0 {
		t.Errorf("expected 0 chunks after Clear, got %d", len(g.chunks))
	}
}

// AddChunks: when growing by exactly one chunk and a spare (from a previous
// trim) is available, it reuses that spare's backing array instead of
// allocating fresh memory.
func TestAddChunks_ReusesSpareAfterTrim(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.Extend(0, cap) // fill chunk 0
	g.AllocSlot()    // spills into chunk 1
	spareBackingPtr := g.chunks[1].Ptr

	g.FreeSlot(1)   // empty chunk 1
	g.ResolveTail() // trims chunk 1, stashing its backing array as the spare

	if len(g.chunks) != 1 {
		t.Fatalf("expected chunk 1 to be trimmed, got %d chunks", len(g.chunks))
	}

	g.AllocSlot() // chunk 0 is full again — must grow, reusing the spare

	if len(g.chunks) != 2 {
		t.Fatalf("expected growth to add back a chunk, got %d", len(g.chunks))
	}
	if g.chunks[1].Ptr != spareBackingPtr {
		t.Error("expected AddChunks to reuse the spare chunk's backing array, got a freshly allocated one")
	}
}
