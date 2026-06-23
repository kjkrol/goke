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
type Col[T any] struct {
	Idx int
}

func (c *Col[T]) Slice(cur *Cursor) []T {
	return unsafe.Slice((*T)(unsafe.Add(cur.Base, cur.Offsets[c.Idx])), len(cur.IDs))
}

func (c *Col[T]) At(cur *Cursor) *T {
	var zero T
	return (*T)(unsafe.Add(cur.Base, cur.Offsets[c.Idx]+cur.Slot*unsafe.Sizeof(zero)))
}
