package mem

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func newTestBlock(t *testing.T) Block {
	t.Helper()
	var b Block
	var layout ChunkLayout
	layout.Init([]comp.Meta{{ID: 1, Size: 8, Align: 8}})
	b.Init(layout)
	return b
}

// PrepareSlots: room in last chunk, count fits entirely → no new chunks
func TestPrepareSlots_FitsInCurrent(t *testing.T) {
	b := newTestBlock(t)
	cap := int(b.Layout.ChunkCap)

	b.AllocSlots(0, 1) // occupy one slot

	startIdx, available := b.PrepareSlots(1)

	if startIdx != 0 {
		t.Errorf("expected startIdx 0, got %d", startIdx)
	}
	if available != cap-1 {
		t.Errorf("expected available %d, got %d", cap-1, available)
	}
	if b.NumChunks() != 1 {
		t.Errorf("expected 1 chunk, got %d", b.NumChunks())
	}
	if b.Reserved != 0 {
		t.Errorf("expected Reserved 0, got %d", b.Reserved)
	}
}

// PrepareSlots: room in last chunk, count spills over → extra chunks allocated
func TestPrepareSlots_SpillsOverCurrent(t *testing.T) {
	b := newTestBlock(t)
	cap := int(b.Layout.ChunkCap)

	b.AllocSlots(0, 1) // occupy one slot, available = cap-1
	count := cap + 1   // needs current partial + at least one full new chunk

	startIdx, available := b.PrepareSlots(count)

	if startIdx != 0 {
		t.Errorf("expected startIdx 0, got %d", startIdx)
	}
	if available != cap-1 {
		t.Errorf("expected available %d, got %d", cap-1, available)
	}
	if b.NumChunks() < 2 {
		t.Errorf("expected at least 2 chunks, got %d", b.NumChunks())
	}
	if int(b.Reserved) != b.NumChunks()-1 {
		t.Errorf("expected Reserved %d, got %d", b.NumChunks()-1, b.Reserved)
	}
}

// PrepareSlots: last chunk is full → starts from a fresh chunk
func TestPrepareSlots_LastChunkFull(t *testing.T) {
	b := newTestBlock(t)
	cap := int(b.Layout.ChunkCap)

	b.AllocSlots(0, cap) // fill chunk 0 completely

	startIdx, available := b.PrepareSlots(1)

	if startIdx != 1 {
		t.Errorf("expected startIdx 1, got %d", startIdx)
	}
	if available != cap {
		t.Errorf("expected available %d, got %d", cap, available)
	}
	if b.NumChunks() < 2 {
		t.Errorf("expected at least 2 chunks, got %d", b.NumChunks())
	}
}

// PrepareSlots: Reserved is always set to the last allocated chunk index
func TestPrepareSlots_ReservedTracksLastChunk(t *testing.T) {
	b := newTestBlock(t)
	cap := int(b.Layout.ChunkCap)

	b.PrepareSlots(cap * 3)

	if int(b.Reserved) != b.NumChunks()-1 {
		t.Errorf("Reserved %d != last chunk index %d", b.Reserved, b.NumChunks()-1)
	}
}
