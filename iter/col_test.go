package iter

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"
)

func TestCol_Slice(t *testing.T) {
	buf := make([]byte, 64)
	cur := &Cursor{
		Base:    unsafe.Pointer(&buf[0]),
		Offsets: []uintptr{0},
		IDs:     make([]uid.UID64, 3),
	}

	var col Col[int32]
	s := col.Slice(cur)
	if len(s) != 3 {
		t.Fatalf("expected slice length 3 (== len(cur.IDs)), got %d", len(s))
	}

	s[0], s[1], s[2] = 10, 20, 30

	// A fresh Slice() call must see the same underlying memory, not a copy.
	again := col.Slice(cur)
	if again[0] != 10 || again[1] != 20 || again[2] != 30 {
		t.Errorf("expected writes to persist through the shared memory, got %v", again)
	}
}

func TestCol_Slice_RespectsOffsetAndIdx(t *testing.T) {
	buf := make([]byte, 64)
	cur := &Cursor{
		Base:    unsafe.Pointer(&buf[0]),
		Offsets: []uintptr{0, 16}, // colA's column at byte 0, colB's at byte 16
		IDs:     make([]uid.UID64, 2),
	}

	var colA Col[int32]
	colA.Idx = 0
	var colB Col[int64]
	colB.Idx = 1

	sa := colA.Slice(cur)
	sb := colB.Slice(cur)
	sa[0], sa[1] = 1, 2
	sb[0], sb[1] = 100, 200

	if sa[0] != 1 || sa[1] != 2 {
		t.Errorf("colA: expected [1 2], got %v", sa)
	}
	if sb[0] != 100 || sb[1] != 200 {
		t.Errorf("colB: expected [100 200], got %v", sb)
	}
}

func TestCol_At(t *testing.T) {
	buf := make([]byte, 64)
	cur := &Cursor{
		Base:    unsafe.Pointer(&buf[0]),
		Offsets: []uintptr{0},
	}

	var col Col[int32]
	cur.Slot = 2
	*col.At(cur) = 42

	// Confirm it landed at slot 2, not slot 0 — read back via Slice over the
	// same memory.
	cur.IDs = make([]uid.UID64, 5)
	s := col.Slice(cur)
	if s[2] != 42 {
		t.Errorf("expected slot 2 to be 42, got %v (full slice: %v)", s[2], s)
	}
	for i, v := range s {
		if i != 2 && v != 0 {
			t.Errorf("expected slot %d to be untouched (0), got %v", i, v)
		}
	}
}

func TestCol_At_DifferentSlots(t *testing.T) {
	buf := make([]byte, 64)
	cur := &Cursor{
		Base:    unsafe.Pointer(&buf[0]),
		Offsets: []uintptr{0},
	}

	var col Col[int32]
	for slot := uintptr(0); slot < 4; slot++ {
		cur.Slot = slot
		*col.At(cur) = int32(slot * 10)
	}

	for slot := uintptr(0); slot < 4; slot++ {
		cur.Slot = slot
		want := int32(slot * 10)
		if got := *col.At(cur); got != want {
			t.Errorf("slot %d: expected %d, got %d", slot, want, got)
		}
	}
}

func TestCol_At_RespectsIdx(t *testing.T) {
	buf := make([]byte, 64)
	cur := &Cursor{
		Base:    unsafe.Pointer(&buf[0]),
		Offsets: []uintptr{0, 8},
		Slot:    1,
	}

	var colA Col[int32]
	colA.Idx = 0
	var colB Col[int32]
	colB.Idx = 1

	*colA.At(cur) = 7
	*colB.At(cur) = 9

	if got := *colA.At(cur); got != 7 {
		t.Errorf("colA: expected 7, got %d", got)
	}
	if got := *colB.At(cur); got != 9 {
		t.Errorf("colB: expected 9, got %d", got)
	}
}
