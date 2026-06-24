package iter

import (
	"unsafe"

	"github.com/kjkrol/uid"
)

// Cursor holds the current iteration position. Pass a pointer to it into
// ArrayRef[T].Slice or ArrayRef[T].At.
type Cursor struct {
	Base    unsafe.Pointer
	Offsets []uintptr
	Slot    uintptr
	IDs     []uid.UID64
}
