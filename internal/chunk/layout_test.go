package chunk

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

func TestLayout_Init_FitsWithinL1(t *testing.T) {
	var l Layout
	l.Init([]comp.Def{
		{ID: 1, Size: 8, Align: 8},
		{ID: 2, Size: 4, Align: 4},
	})

	if l.ChunkCap == 0 {
		t.Fatal("expected a non-zero ChunkCap")
	}
	if l.ChunkBytes > L1DataCacheSize {
		t.Errorf("expected ChunkBytes (%d) to fit within L1DataCacheSize (%d)", l.ChunkBytes, L1DataCacheSize)
	}
	if len(l.Offsets) != 3 { // entity column + 2 components
		t.Errorf("expected 3 offsets, got %d", len(l.Offsets))
	}
}

// A component larger than the whole L1 cache can never fit more than one
// row per chunk — Init must fall back to ChunkCap=1 instead of looping
// forever or computing a zero capacity.
func TestLayout_Init_HugeComponentForcesCapacityOne(t *testing.T) {
	var l Layout
	l.Init([]comp.Def{
		{ID: 1, Size: L1DataCacheSize * 2, Align: 8},
	})

	if l.ChunkCap != 1 {
		t.Errorf("expected ChunkCap 1 for an oversized component, got %d", l.ChunkCap)
	}
}
