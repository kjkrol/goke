package iter

import (
	"unsafe"

	"github.com/kjkrol/uid"
)

// Cursor holds the current iteration position. Pass a pointer to it into
// ArrayRef[T].Slice or ArrayRef[T].At.
type Cursor struct {
	Base       unsafe.Pointer
	Offsets    []uintptr
	OptOffsets []uintptr
	OptPresent []bool
	Slot       uintptr
	IDs        []uid.UID64
}

// Set positions the cursor at (base, slot) without touching Offsets or IDs.
func (c *Cursor) Set(base unsafe.Pointer, slot uintptr) {
	c.Base = base
	c.Slot = slot
}
