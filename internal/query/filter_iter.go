package query

// nextFilter advances the View's Filter-mode iterator to the next matching entity.
// Sets v.Entity, v.Idx, and v.Cursor for the matched entity. Returns false when exhausted.
func (v *View) nextFilter() bool {
	for v.pos < len(v.selected) {
		e := v.selected[v.pos]
		v.Idx = v.pos
		v.pos++
		link, ok := v.EntityIndex.Get(e)
		if !ok {
			continue
		}
		if link.ArchId != v.lastArchID {
			v.bt = v.Get(link.ArchId)
			v.lastArchID = link.ArchId
			if v.bt != nil {
				v.Cursor.Offsets = v.bt.CompOffsets // set once per archetype change
			}
		}
		if v.bt == nil {
			continue
		}
		v.Entity = e
		v.bt.PointCursor(&v.Cursor, link.Pos) // per entity: chunk base + slot only
		return true
	}
	return false
}
