package comp

// Composition describes the component composition of an archetype:
// a Mask (bitset of IDs) and the metadata for each non-tag component.
// Immutable after construction — use With and Without to derive new instances.
type Composition struct {
	Mask  Mask
	Metas []Meta
}

// With returns a new Composition with compMeta added.
// Tags (Size == 0) update only the mask, not Metas.
func (s Composition) With(compMeta Meta) Composition {
	newMask := s.Mask.Set(compMeta.ID)
	if compMeta.Size == 0 {
		return Composition{Mask: newMask, Metas: s.Metas}
	}
	newMetas := make([]Meta, len(s.Metas)+1)
	copy(newMetas, s.Metas)
	newMetas[len(s.Metas)] = compMeta
	return Composition{Mask: newMask, Metas: newMetas}
}

// Without returns a new Composition with compID removed.
// Returns s unchanged (no allocation) if compID is not set in the mask.
func (s Composition) Without(compID ID) Composition {
	if !s.Mask.IsSet(compID) {
		return s
	}
	newMask := s.Mask.Clear(compID)
	newMetas := make([]Meta, 0, len(s.Metas)-1)
	for _, m := range s.Metas {
		if m.ID != compID {
			newMetas = append(newMetas, m)
		}
	}
	return Composition{Mask: newMask, Metas: newMetas}
}
