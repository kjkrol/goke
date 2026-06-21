package comp

// Composition describes the component composition of an archetype:
// a Mask (bitset of IDs) and the metadata for each non-tag component.
// Immutable after construction — use With and Without to derive new instances.
type Composition struct {
	Mask Mask
	Defs []Def
}

// With returns a new Composition with compDef added.
// Tags (Size == 0) update only the mask, not Defs.
func (s Composition) With(compDef Def) Composition {
	newMask := s.Mask.Set(compDef.ID)
	if compDef.Size == 0 {
		return Composition{Mask: newMask, Defs: s.Defs}
	}
	newDefs := make([]Def, len(s.Defs)+1)
	copy(newDefs, s.Defs)
	newDefs[len(s.Defs)] = compDef
	return Composition{Mask: newMask, Defs: newDefs}
}

// Without returns a new Composition with compID removed.
// Returns s unchanged (no allocation) if compID is not set in the mask.
func (s Composition) Without(compID ID) Composition {
	if !s.Mask.IsSet(compID) {
		return s
	}
	newMask := s.Mask.Clear(compID)
	newMetas := make([]Def, 0, len(s.Defs)-1)
	for _, m := range s.Defs {
		if m.ID != compID {
			newMetas = append(newMetas, m)
		}
	}
	return Composition{Mask: newMask, Defs: newMetas}
}
