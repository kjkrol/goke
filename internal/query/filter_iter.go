package query

// nextPick advances the Pick-mode walk to the next matching entity.
// Sets w.Entity, w.Idx, and w.Cursor for the matched entity. Returns false when exhausted.
func (w *Walker) nextPick() bool {
	for w.pos < len(w.selected) {
		e := w.selected[w.pos]
		w.Idx = w.pos
		w.pos++
		link, ok := w.entityIndex.Get(e)
		if !ok {
			continue
		}
		if link.ArchID != w.lastArchID {
			w.bt = w.tables.Get(link.ArchID)
			w.lastArchID = link.ArchID
			if w.bt != nil {
				w.Cursor.Offsets = w.bt.CompOffsets
				w.Cursor.OptOffsets = w.bt.OptOffsets
				w.Cursor.OptPresent = w.bt.OptPresent
			}
		}
		if w.bt == nil {
			continue
		}
		w.Entity = e
		w.Cursor.Set(link.ChunkPtr, uintptr(link.Slot)) // per entity: chunk base + slot only
		return true
	}
	return false
}
