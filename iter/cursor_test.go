package iter

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"
)

func TestCursor_Set_PositionsBaseAndSlot_LeavesOffsetsAndIDs(t *testing.T) {
	buf := make([]byte, 64)
	origOffsets := []uintptr{4, 8}
	origIDs := []uid.UID64{1, 2, 3}
	cur := &Cursor{
		Base:    unsafe.Pointer(&buf[0]),
		Offsets: origOffsets,
		IDs:     origIDs,
	}

	newBase := unsafe.Pointer(&buf[16])
	cur.Set(newBase, 5)

	if cur.Base != newBase {
		t.Errorf("expected Base = %v, got %v", newBase, cur.Base)
	}
	if cur.Slot != 5 {
		t.Errorf("expected Slot = 5, got %d", cur.Slot)
	}
	if &cur.Offsets[0] != &origOffsets[0] {
		t.Error("expected Set to leave Offsets untouched")
	}
	if &cur.IDs[0] != &origIDs[0] {
		t.Error("expected Set to leave IDs untouched")
	}
}
