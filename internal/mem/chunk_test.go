package mem

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func TestBlock_ChunkAccessors(t *testing.T) {
	var b Block
	layout := ChunkLayout{}
	layout.Init([]comp.Meta{{ID: 1, Size: 8, Align: 8}})
	b.Init(layout)

	if b.NumChunks() != 1 {
		t.Fatalf("expected 1 chunk after Init, got %d", b.NumChunks())
	}
	if b.ChunkLen(0) != 0 {
		t.Errorf("expected ChunkLen(0) == 0, got %d", b.ChunkLen(0))
	}
	if b.ChunkPtr(0) == nil {
		t.Error("expected non-nil ChunkPtr(0)")
	}
}

func TestBlock_AllocSlotsAndFreeSlot(t *testing.T) {
	var b Block
	layout := ChunkLayout{}
	layout.Init([]comp.Meta{{ID: 1, Size: 8, Align: 8}})
	b.Init(layout)

	b.AllocSlots(0, 3)
	if b.ChunkLen(0) != 3 {
		t.Errorf("expected ChunkLen 3 after Advance(3), got %d", b.ChunkLen(0))
	}

	b.FreeSlot(0)
	if b.ChunkLen(0) != 2 {
		t.Errorf("expected ChunkLen 2 after Decr, got %d", b.ChunkLen(0))
	}
}
