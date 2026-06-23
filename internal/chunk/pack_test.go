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
