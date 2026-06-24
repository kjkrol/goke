package colstore

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

func TestTable_BakeColumnsAndOffsets(t *testing.T) {
	defs := []comp.Def{
		{ID: 1, Size: 8, Align: 8},
		{ID: 2, Size: 8, Align: 8},
	}
	var tbl Table
	tbl.Init(defs)

	baked := tbl.BakeColumns(defs)
	if len(baked) != 2 {
		t.Fatalf("expected 2 baked columns, got %d", len(baked))
	}
	for i, b := range baked {
		if b.CompSize != defs[i].Size {
			t.Errorf("comp %d: expected CompSize %d, got %d", i, defs[i].Size, b.CompSize)
		}
	}
	if baked[0].Offset == baked[1].Offset {
		t.Error("expected distinct offsets for distinct tracked components")
	}

	// Untracked component ID hits the col == nil branch — leaves the zero value.
	missing := tbl.BakeColumns([]comp.Def{{ID: 99, Size: 8, Align: 8}})
	if missing[0] != (ColBake{}) {
		t.Errorf("expected zero-value ColBake for untracked component, got %+v", missing[0])
	}

	offsets := tbl.BakeOffsets([]comp.ID{1, 2})
	if offsets[0] != baked[0].Offset || offsets[1] != baked[1].Offset {
		t.Errorf("BakeOffsets %v doesn't match BakeColumns offsets %d/%d", offsets, baked[0].Offset, baked[1].Offset)
	}

	missingOffsets := tbl.BakeOffsets([]comp.ID{99})
	if missingOffsets[0] != 0 {
		t.Errorf("expected offset 0 for untracked component, got %d", missingOffsets[0])
	}
}

func TestTable_ComponentAt(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)

	cur := newCursor(1)
	_, pos := tbl.SpawnCursor(cur, 0, 1, baked)

	ptr := tbl.ComponentAt(pos, 1)
	if ptr == nil {
		t.Fatal("expected non-nil pointer for tracked component")
	}
	*(*int64)(ptr) = 42
	if got := *(*int64)(tbl.ComponentAt(pos, 1)); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}

	if tbl.ComponentAt(pos, 99) != nil {
		t.Error("expected nil pointer for an untracked component ID")
	}
}
