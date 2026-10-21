package query

// nextAll advances the All-mode walk to the next non-empty chunk.
// Populates w.Cursor. Returns false when exhausted.
func (w *Walker) nextAll() bool {
	w.chunkIdx++
	for w.tableIdx < len(w.tables.BakedTables) {
		bt := &w.tables.BakedTables[w.tableIdx]
		if idx, ok := bt.FillCursorNext(&w.Cursor, w.chunkIdx); ok {
			w.chunkIdx = idx
			return true
		}
		w.tableIdx++
		w.chunkIdx = 0
	}
	w.Cursor.IDs = nil
	return false
}

// nextPaced is nextAll over the tables whose turn it is.
func (w *Walker) nextPaced() bool {
	w.chunkIdx++
	for w.tableIdx < len(w.tables.BakedTables) {
		bt := &w.tables.BakedTables[w.tableIdx]
		if w.due(bt) {
			if idx, ok := bt.FillCursorNext(&w.Cursor, w.chunkIdx); ok {
				w.chunkIdx = idx
				return true
			}
		}
		w.tableIdx++
		w.chunkIdx = 0
	}
	w.Cursor.IDs = nil
	return false
}
