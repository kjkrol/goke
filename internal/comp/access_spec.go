package comp

import (
	"fmt"
	"slices"
)

type AccessSpec struct {
	CompInfos    []Def
	OptCompInfos []Def
	TagIDs       []ID
	ExCompIDs    []ID
	Pace         *Pace
	Reads        Mask
}

// Access is what queries do to components through their tracked columns: Writes those they may
// write, Reads those they only read.
type Access struct{ Reads, Writes Mask }

// Access is what s's tracked columns do; a tag, holding no data, is neither.
func (s *AccessSpec) Access() Access {
	a := Access{Reads: s.Reads}
	for _, defs := range [2][]Def{s.CompInfos, s.OptCompInfos} {
		for _, def := range defs {
			if def.Size > 0 && !s.Reads.IsSet(def.ID) {
				a.Writes = a.Writes.Set(def.ID)
			}
		}
	}
	return a
}

// Add has a do what b does too.
func (a *Access) Add(b Access) {
	for i := range MaskSize {
		a.Reads[i] |= b.Reads[i]
		a.Writes[i] |= b.Writes[i]
	}
}

// Conflicts are the components one of a and b writes while the other reads or writes them.
func (a Access) Conflicts(b Access) Mask {
	var m Mask
	for i := range MaskSize {
		m[i] = a.Writes[i]&(b.Writes[i]|b.Reads[i]) | b.Writes[i]&a.Reads[i]
	}
	return m
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

// OptCompIDs returns the IDs of the optionally-tracked data columns in track order.
func (s *AccessSpec) OptCompIDs() []ID {
	ids := make([]ID, len(s.OptCompInfos))
	for i, def := range s.OptCompInfos {
		ids[i] = def.ID
	}
	return ids
}

// OptComp registers def as a data column that never gates a match — a
// zero-size def (tag) is tracked via the archetype mask instead of a column.
func (s *AccessSpec) OptComp(def Def) error {
	for _, existing := range s.OptCompInfos {
		if existing.ID == def.ID {
			return fmt.Errorf("component %s (ID: %d) is already optional in this access spec", def.Type.String(), def.ID)
		}
	}
	for _, existing := range s.CompInfos {
		if existing.ID == def.ID {
			return fmt.Errorf("component %s (ID: %d) is already required in this access spec, cannot also be optional", def.Type.String(), def.ID)
		}
	}
	if slices.Contains(s.TagIDs, def.ID) {
		return fmt.Errorf("component %s (ID: %d) is already required in this access spec, cannot also be optional", def.Type.String(), def.ID)
	}
	s.OptCompInfos = append(s.OptCompInfos, def)
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
