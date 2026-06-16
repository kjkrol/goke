package query

import (
	"unsafe"

	"github.com/kjkrol/uid"
)

// nextAll advances the View's All-mode iterator to the next non-empty chunk.
// Stores the chunk's entity column in v.EntSlice. Returns false when exhausted.
func (v *View) nextAll() bool {
	v.chunkIdx++
	for v.tableIdx < len(v.BakedTables) {
		bt := &v.BakedTables[v.tableIdx]
		chunks := bt.Table.Chunks
		for v.chunkIdx < len(chunks) {
			chunk := &chunks[v.chunkIdx]
			if chunk.Len > 0 {
				v.EntSlice = unsafe.Slice((*uid.UID64)(chunk.Ptr), chunk.Len)
				return true
			}
			v.chunkIdx++
		}
		v.tableIdx++
		v.chunkIdx = 0
	}
	v.EntSlice = nil
	return false
}

// slice returns a typed slice for the component at compIdx in the current chunk.
// Storing v.EntSlice lets the compiler prove i < len(slice) and eliminate bounds
// checks on component indexing when ranging v.EntSlice in the same loop.
func slice[T any](v *View, compIdx int) []T {
	base := unsafe.Pointer(unsafe.SliceData(v.EntSlice))
	return unsafe.Slice((*T)(unsafe.Add(base, v.BakedTables[v.tableIdx].CompOffsets[compIdx])), len(v.EntSlice))
}
