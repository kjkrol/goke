package iter

import "unsafe"

// ArrayRef locates the array for component type T within a Cursor's data.
//
// KNOWN ISSUE: cur.Base points into a chunk allocated as a raw []byte (see
// internal/chunk), which the garbage collector marks noscan at allocation
// time and never scans for pointers, regardless of how the memory is later
// reinterpreted here. If T contains a Go pointer (a string's backing array,
// a BinaryMarshaler-backed type with a pointer field, ...), that pointer is
// invisible to the GC — its target can be collected while still referenced
// only from chunk memory, corrupting the field. Currently affects any
// component with a string or pointer-bearing field (comp.ValidateEncodable
// permits both). Not yet fixed; see internal/comp/doc.go's Encoding
// constraints section.
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
