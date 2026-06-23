package chunk

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
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
