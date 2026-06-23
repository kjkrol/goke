package query

// nextAll advances the Matcher's All-mode iterator to the next non-empty chunk.
// Populates m.Cursor. Returns false when exhausted.
func (m *Matcher) nextAll() bool {
	m.chunkIdx++
	for m.tableIdx < len(m.BakedTables) {
		bt := &m.BakedTables[m.tableIdx]
		if idx, ok := bt.FillCursorNext(&m.Cursor, m.chunkIdx); ok {
			m.chunkIdx = idx
			return true
		}
		m.tableIdx++
		m.chunkIdx = 0
	}
	m.Cursor.IDs = nil
	return false
}
