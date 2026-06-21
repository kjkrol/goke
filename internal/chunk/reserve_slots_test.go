package chunk

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func newTestPack(t *testing.T) Pack {
	t.Helper()
	var g Pack
	var layout Layout
	layout.Init([]comp.Def{{ID: 1, Size: 8, Align: 8}})
	g.Init(layout)
	return g
}

// ReserveSlots: room in last chunk, count fits entirely → no new chunks
func TestReserveSlots_FitsInCurrent(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.commitSlots(0, 1) // occupy one slot

	startIdx, available := g.ReserveSlots(1)

	if startIdx != 0 {
		t.Errorf("expected startIdx 0, got %d", startIdx)
	}
	if available != cap-1 {
		t.Errorf("expected available %d, got %d", cap-1, available)
	}
	if g.NumChunks() != 1 {
		t.Errorf("expected 1 chunk, got %d", g.NumChunks())
	}
	if g.Reserved != 0 {
		t.Errorf("expected Reserved 0, got %d", g.Reserved)
	}
}

// ReserveSlots: room in last chunk, count spills over → extra chunks allocated
func TestReserveSlots_SpillsOverCurrent(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.commitSlots(0, 1) // occupy one slot, available = cap-1
	count := cap + 1    // needs current partial + at least one full new chunk

	startIdx, available := g.ReserveSlots(count)

	if startIdx != 0 {
		t.Errorf("expected startIdx 0, got %d", startIdx)
	}
	if available != cap-1 {
		t.Errorf("expected available %d, got %d", cap-1, available)
	}
	if g.NumChunks() < 2 {
		t.Errorf("expected at least 2 chunks, got %d", g.NumChunks())
	}
	if int(g.Reserved) != g.NumChunks()-1 {
		t.Errorf("expected Reserved %d, got %d", g.NumChunks()-1, g.Reserved)
	}
}

// ReserveSlots: last chunk is full → starts from a fresh chunk
func TestReserveSlots_LastChunkFull(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.commitSlots(0, cap) // fill chunk 0 completely

	startIdx, available := g.ReserveSlots(1)

	if startIdx != 1 {
		t.Errorf("expected startIdx 1, got %d", startIdx)
	}
	if available != cap {
		t.Errorf("expected available %d, got %d", cap, available)
	}
	if g.NumChunks() < 2 {
		t.Errorf("expected at least 2 chunks, got %d", g.NumChunks())
	}
}

// ReserveSlots: Reserved is always set to the last allocated chunk index
func TestReserveSlots_ReservedTracksLastChunk(t *testing.T) {
	g := newTestPack(t)
	cap := int(g.Layout.ChunkCap)

	g.ReserveSlots(cap * 3)

	if int(g.Reserved) != g.NumChunks()-1 {
		t.Errorf("Reserved %d != last chunk index %d", g.Reserved, g.NumChunks()-1)
	}
}
