package iter

import "unsafe"

// ArrayRef locates the array for component type T within a Cursor's data.
//
// cur.Base points into a chunk (see internal/chunk); when T has a field a
// GC pointer could hide in (a string, a BinaryMarshaler-backed value), that
// chunk was allocated as GC-scanned memory rather than a raw noscan []byte,
// so the pointer stays reachable across GC cycles regardless of how it's
// reinterpreted here.
type ArrayRef[T any] struct {
	Idx int
}

// Slice returns the array of T currently addressable via cur.Base/Offsets.
func (c *ArrayRef[T]) Slice(cur *Cursor) []T {
	return unsafe.Slice((*T)(unsafe.Add(cur.Base, cur.Offsets[c.Idx])), len(cur.IDs))
}

// At returns a pointer to the array element at cur.Slot.
func (c *ArrayRef[T]) At(cur *Cursor) *T {
	var zero T
	return (*T)(unsafe.Add(cur.Base, cur.Offsets[c.Idx]+cur.Slot*unsafe.Sizeof(zero)))
}
