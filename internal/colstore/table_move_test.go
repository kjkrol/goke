package colstore

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

// MoveEntityFrom: dst gains only the columns it shares with src; the rest of
// dst's columns (newly added by the caller, not present in src) are left
// untouched for the caller to fill in afterwards.
func TestTable_MoveEntityFrom_CopiesOnlySharedColumns(t *testing.T) {
	srcDefs := []comp.Def{{ID: 1, Size: 8, Align: 8}, {ID: 2, Size: 8, Align: 8}}
	dstDefs := []comp.Def{{ID: 2, Size: 8, Align: 8}, {ID: 3, Size: 8, Align: 8}}

	src := newTestTable(t, srcDefs)
	dst := newTestTable(t, dstDefs)

	srcBaked := src.BakeColumns(srcDefs)
	cur := newCursor(len(srcDefs))
	_, srcPos := src.SpawnCursor(cur, 0, 1, srcBaked)
	entityID := cur.IDs[0]

	*(*int64)(src.ComponentAt(srcPos, 1)) = 111
	*(*int64)(src.ComponentAt(srcPos, 2)) = 222

	newPos, swappedEntity, swapped := dst.MoveEntityFrom(src, entityID, srcPos)

	if swapped {
		t.Errorf("expected no swap (only entity in src), got swapped=%v entity=%v", swapped, swappedEntity)
	}
	if src.Len() != 0 {
		t.Errorf("expected src to be empty after the move, got Len %d", src.Len())
	}
	if dst.Len() != 1 {
		t.Errorf("expected dst to have 1 entity after the move, got Len %d", dst.Len())
	}

	if got := *(*int64)(dst.ComponentAt(newPos, 2)); got != 222 {
		t.Errorf("expected shared comp2 to be copied as 222, got %d", got)
	}
	ptr := dst.ComponentAt(newPos, 3)
	if ptr == nil {
		t.Fatal("expected dst's comp3 column to exist")
	}
	if got := *(*int64)(ptr); got != 0 {
		t.Errorf("expected comp3 (not present in src) to stay zero, got %d", got)
	}
}

// MoveEntityFrom removing a non-last slot from src must swap the last
// entity into the vacated slot to keep src dense.
func TestTable_MoveEntityFrom_SwapsLastEntityIntoHole(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8}}
	src := newTestTable(t, defs)
	dst := newTestTable(t, defs)

	baked := src.BakeColumns(defs)
	cur := newCursor(1)

	_, pos0 := src.SpawnCursor(cur, 0, 1, baked)
	id0 := cur.IDs[0]
	_, pos1 := src.SpawnCursor(cur, 0, 1, baked)
	id1 := cur.IDs[0]

	*(*int64)(src.ComponentAt(pos0, 1)) = 10
	*(*int64)(src.ComponentAt(pos1, 1)) = 20

	// Move the first entity (pos0) out — the last one (id1, at pos1) must
	// swap into pos0's hole.
	_, swappedEntity, swapped := dst.MoveEntityFrom(src, id0, pos0)

	if !swapped {
		t.Fatal("expected a swap since pos0 wasn't the last slot")
	}
	if swappedEntity != id1 {
		t.Errorf("expected swapped entity to be id1 (%v), got %v", id1, swappedEntity)
	}
	if src.Len() != 1 {
		t.Errorf("expected src to have 1 entity left, got %d", src.Len())
	}
	if got := *(*int64)(src.ComponentAt(pos0, 1)); got != 20 {
		t.Errorf("expected id1's value (20) to be swapped into pos0, got %d", got)
	}
}
