package query

// nextAll advances the View's All-mode iterator to the next non-empty chunk.
// Populates v.Cursor. Returns false when exhausted.
func (v *View) nextAll() bool {
	v.chunkIdx++
	for v.tableIdx < len(v.BakedTables) {
		bt := &v.BakedTables[v.tableIdx]
		if idx, ok := bt.FillCursorNext(&v.Cursor, v.chunkIdx); ok {
			v.chunkIdx = idx
			return true
		}
		v.tableIdx++
		v.chunkIdx = 0
	}
	v.Cursor.IDs = nil
	return false
}
