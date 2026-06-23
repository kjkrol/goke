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
		if link.ArchId != m.lastArchID {
			m.bt = m.Get(link.ArchId)
			m.lastArchID = link.ArchId
			if m.bt != nil {
				m.Cursor.Offsets = m.bt.CompOffsets // set once per archetype change
			}
		}
		if m.bt == nil {
			continue
		}
		m.Entity = e
		m.bt.PointCursor(&m.Cursor, link.Pos) // per entity: chunk base + slot only
		return true
	}
	return false
}
