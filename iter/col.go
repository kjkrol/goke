package iter

import (
	"unsafe"

	"github.com/kjkrol/uid"
)

// Cursor holds the current iteration position and the data needed to access
// component columns. Pass a pointer to it into Col[T].Slice or Col[T].At.
//
// IDs is the canonical entity slice for the current batch.
// Range over cursor.IDs to let the compiler prove i < len(col.Slice(cursor))
// and eliminate bounds checks for any number of tracked columns.
type Cursor struct {
	Base    unsafe.Pointer
	Offsets []uintptr
	Slot    uintptr
	IDs     []uid.UID64
}

// Col is a typed column handle for a tracked component.
// Declare one as a struct field, register it via Track(&col), then use
// col.Slice(cursor) in All/Factory mode and col.At(cursor) in Filter mode.
type Col[T any] struct {
	Idx int
}

// Slice returns the component slice for the current All-mode chunk or Factory batch.
// Its length equals len(cursor.IDs), so ranging cursor.IDs in the inner
// loop lets the compiler eliminate bounds checks on slice[i] accesses.
func (c *Col[T]) Slice(cur *Cursor) []T {
	return unsafe.Slice((*T)(unsafe.Add(cur.Base, cur.Offsets[c.Idx])), len(cur.IDs))
}

// At returns a pointer to the component for the current Filter-mode entity.
func (c *Col[T]) At(cur *Cursor) *T {
	var zero T
	return (*T)(unsafe.Add(cur.Base, cur.Offsets[c.Idx]+cur.Slot*unsafe.Sizeof(zero)))
}
