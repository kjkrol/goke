package query

import (
	"unsafe"

	"github.com/kjkrol/uid"
)

// nextAll advances the View's All-mode iterator to the next non-empty chunk.
// Populates v.Cursor. Returns false when exhausted.
func (v *View) nextAll() bool {
	v.chunkIdx++
	for v.tableIdx < len(v.BakedTables) {
		bt := &v.BakedTables[v.tableIdx]
		if idx, ptr, chunkLen, ok := bt.Table.NextNonEmptyChunk(v.chunkIdx); ok {
			v.chunkIdx = idx
			v.Cursor.EntSlice = unsafe.Slice((*uid.UID64)(ptr), chunkLen)
			v.Cursor.Base = ptr
			v.Cursor.Offsets = bt.CompOffsets
			return true
		}
		v.tableIdx++
		v.chunkIdx = 0
	}
	v.Cursor.EntSlice = nil
	return false
}
