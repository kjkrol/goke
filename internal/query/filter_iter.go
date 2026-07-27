package query

// nextPick advances the Matcher's Pick-mode iterator to the next matching entity.
// Sets m.Entity, m.Idx, and m.Cursor for the matched entity. Returns false when exhausted.
func (m *Matcher) nextPick() bool {
	for m.pos < len(m.selected) {
		e := m.selected[m.pos]
		m.Idx = m.pos
		m.pos++
		link, ok := m.EntityIndex.Get(e)
		if !ok {
			continue
		}
		if link.ArchID != m.lastArchID {
			m.bt = m.Get(link.ArchID)
			m.lastArchID = link.ArchID
			if m.bt != nil {
				m.Cursor.Offsets = m.bt.CompOffsets // set once per archetype change
			}
		}
		if m.bt == nil {
			continue
		}
		m.Entity = e
		m.Cursor.Set(link.ChunkPtr, uintptr(link.Slot)) // per entity: chunk base + slot only
		return true
	}
	return false
}
