package comp

import (
	"fmt"
	"slices"
)

type AccessSpec struct {
	CompInfos []Def
	TagIDs    []ID
	ExCompIDs []ID
}

// Init applies opts against mi, populating s in place.
// Panics if any opt returns an error.
func (s *AccessSpec) Init(mi *DefIndex, opts ...AccessOpt) {
	for _, opt := range opts {
		if err := opt(s, mi); err != nil {
			panic(fmt.Sprintf("comp: access-spec option: %v", err))
		}
	}
}

// Compose derives a Composition from the AccessSpec without requiring a DefIndex.
func (s *AccessSpec) Compose() Composition {
	return Composition{Mask: NewMask(s), Defs: s.CompInfos}
}

// CompIDs returns the IDs of the tracked data columns in track order.
// Read-only consumers (views, lookups) need only the IDs, not the full Defs.
func (s *AccessSpec) CompIDs() []ID {
	ids := make([]ID, len(s.CompInfos))
	for i, def := range s.CompInfos {
		ids[i] = def.ID
	}
	return ids
}

func (s *AccessSpec) Comp(def Def) error {
	for _, existing := range s.CompInfos {
		if existing.ID == def.ID {
			return fmt.Errorf("component %s (ID: %d) is already in this access spec", def.Type.String(), def.ID)
		}
	}
	if def.Size == 0 {
		return fmt.Errorf("cannot add %s: tags are not allowed as data columns", def.Type.String())
	}
	s.CompInfos = append(s.CompInfos, def)
	return nil
}

func (s *AccessSpec) Tag(tagId ID) error {
	if slices.Contains(s.TagIDs, tagId) {
		return fmt.Errorf("tag with ID %d is already in this access spec", tagId)
	}
	s.TagIDs = append(s.TagIDs, tagId)
	return nil
}

func (s *AccessSpec) Exclude(id ID) error {
	if slices.Contains(s.ExCompIDs, id) {
		return fmt.Errorf("component ID %d is already in the exclusion list", id)
	}
	s.ExCompIDs = append(s.ExCompIDs, id)
	return nil
}
