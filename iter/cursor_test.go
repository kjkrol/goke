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

// Ver follows Base: a Set into another chunk of the same layout points it at that chunk's counter
func TestCursor_Ver_FollowsBase(t *testing.T) {
	var a, b [4]uint32
	cur := &Cursor{Base: unsafe.Pointer(&a[0]), VerOffset: 3 * unsafe.Sizeof(uint32(0))}
	if cur.Ver() != &a[3] {
		t.Error("Ver is not Base plus VerOffset")
	}
	cur.Set(unsafe.Pointer(&b[0]), 1)
	*cur.Ver()++
	if b[3] != 1 {
		t.Errorf("a count through Ver after Set reached %d, want b's counter at 1", b[3])
	}
}
