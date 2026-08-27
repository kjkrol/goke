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

// OptArrayRef locates the array for an optionally-present component type T
// within a Cursor's data — present iff cur.OptPresent[Idx] is true.
type OptArrayRef[T any] struct {
	Idx int
}

// Present reports whether T exists on the cursor's current chunk/entity.
func (c *OptArrayRef[T]) Present(cur *Cursor) bool { return cur.OptPresent[c.Idx] }

// Slice returns T's array for the current chunk, or nil if T is absent.
func (c *OptArrayRef[T]) Slice(cur *Cursor) []T {
	if !cur.OptPresent[c.Idx] {
		return nil
	}
	return unsafe.Slice((*T)(unsafe.Add(cur.Base, cur.OptOffsets[c.Idx])), len(cur.IDs))
}

// At returns a pointer to T at cur.Slot, or nil if T is absent.
func (c *OptArrayRef[T]) At(cur *Cursor) *T {
	if !cur.OptPresent[c.Idx] {
		return nil
	}
	var zero T
	return (*T)(unsafe.Add(cur.Base, cur.OptOffsets[c.Idx]+cur.Slot*unsafe.Sizeof(zero)))
}
