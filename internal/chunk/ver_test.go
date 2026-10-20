package chunk

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/comp"
)

// the counter sits past every column, inside the chunk, on its own aligned word
func TestLayout_VerOffsetPastTheLastColumn(t *testing.T) {
	defs := []comp.Def{
		{ID: 1, Size: 8, Align: 8},
		{ID: 2, Size: 3, Align: 1},
		{ID: 3, Size: 16, Align: 8, Type: reflect.TypeFor[string](), NeedsScan: true},
	}
	var l Layout
	l.Init(defs)

	last := l.Offsets[len(l.Offsets)-1] + uintptr(l.ChunkCap)*defs[len(defs)-1].Size
	if l.VerOffset < last {
		t.Errorf("VerOffset %d lies inside the last column, which ends at %d", l.VerOffset, last)
	}
	if l.VerOffset%unsafe.Alignof(uint32(0)) != 0 {
		t.Errorf("VerOffset %d is not aligned for a uint32", l.VerOffset)
	}
	if l.VerOffset+unsafe.Sizeof(uint32(0)) > l.ChunkBytes {
		t.Errorf("VerOffset %d leaves no room for a counter in %d bytes", l.VerOffset, l.ChunkBytes)
	}
	if f, ok := l.ChunkType.FieldByName("Ver"); !ok || f.Offset != l.VerOffset {
		t.Errorf("the scanned chunk type has no counter at %d: %+v", l.VerOffset, f)
	}
}

// an append counts on its chunk, and only there
func TestPack_AllocSlotAndExtendCountOnTheirChunk(t *testing.T) {
	g := newTestPack(t)
	g.AddChunks(1)

	g.Extend(1, 2)
	g.AllocSlot()
	if got := *g.VerAt(g.ChunkPtr(1)); got != 2 {
		t.Errorf("chunk 1 counts %d after an Extend and an AllocSlot, want 2", got)
	}
	if got := *g.VerAt(g.ChunkPtr(0)); got != 0 {
		t.Errorf("chunk 0 counts %d though nothing was appended to it", got)
	}
}

// the counter lives in the chunk's memory, so a swap and a trim carry it with the chunk
func TestPack_VerMovesWithItsChunk(t *testing.T) {
	g := newTestPack(t)
	chunkCap := int(g.Layout.ChunkCap)
	g.Extend(0, chunkCap)
	g.AllocSlot()
	*g.VerAt(g.ChunkPtr(1)) = 7

	g.SwapChunks(0, 1)
	if got := *g.VerAt(g.ChunkPtr(0)); got != 7 {
		t.Errorf("after the swap chunk 0 counts %d, want the moved chunk's 7", got)
	}
	if got := *g.VerAt(g.ChunkPtr(1)); got != 1 {
		t.Errorf("after the swap chunk 1 counts %d, want the moved chunk's 1", got)
	}

	g.SwapChunks(0, 1)
	g.FreeSlot(1)
	g.ResolveTail()
	if g.NumChunks() != 1 {
		t.Fatalf("expected chunk 1 trimmed, got %d chunks", g.NumChunks())
	}
	g.AllocSlot()
	if got := *g.VerAt(g.ChunkPtr(1)); got != 8 {
		t.Errorf("the reused spare counts %d, want 8: its 7 and the AllocSlot", got)
	}
}
