package colstore

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

// RemoveAt on the last (or only) slot needs no swap, and must zero the
// freed slot's memory.
func TestTable_RemoveAt_LastSlotNoSwap(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	tbl := newTestTable(t, defs)
	baked := tbl.BakeColumns(defs)
	cur := newCursor(1)
	_, pos := tbl.SpawnCursor(cur, 0, 1, baked)

	*(*int64)(tbl.ComponentAt(pos, 1)) = 77

	movedID, swapped := tbl.RemoveAt(pos)

	if swapped {
		t.Errorf("expected no swap removing the only slot, got swapped=%v movedID=%v", swapped, movedID)
	}
	if movedID != 0 {
		t.Errorf("expected movedID 0 when no swap occurs, got %v", movedID)
	}
	if tbl.Len() != 0 {
		t.Errorf("expected Len 0 after removing the only entity, got %d", tbl.Len())
	}
	if got := *(*int64)(tbl.ComponentAt(pos, 1)); got != 0 {
		t.Errorf("expected the freed slot to be zeroed, got %d", got)
	}
}
