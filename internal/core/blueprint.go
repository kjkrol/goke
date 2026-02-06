package core

import "fmt"

type Blueprint struct {
	Reg       *Registry
	compInfos []ComponentInfo
	tagIDs    []ComponentID
	exCompIDs []ComponentID
}

func NewBlueprint(reg *Registry) *Blueprint {
	return &Blueprint{
		Reg:       reg,
		compInfos: make([]ComponentInfo, 0, 8),
		tagIDs:    make([]ComponentID, 0, 8),
		exCompIDs: make([]ComponentID, 0, 8),
	}
}

// WithTag adds a tag ID to the blueprint.
// Returns an error if the tag is already present to ensure definition clarity.
func (b *Blueprint) WithTag(tagId ComponentID) error {
	for _, id := range b.tagIDs {
		if id == tagId {
			return fmt.Errorf("tag with ID %d is already defined in this blueprint", tagId)
		}
	}
	b.tagIDs = append(b.tagIDs, tagId)
	return nil
}

// WithComp adds a data-carrying component.
// Returns error if it's a duplicate or has no data (Size == 0).
func (b *Blueprint) WithComp(info ComponentInfo) error {
	// Ensure the component is not already part of this blueprint to prevent duplicates
	for _, existing := range b.compInfos {
		if existing.ID == info.ID {
			return fmt.Errorf("component %s (ID: %d) is already defined in this blueprint", info.Type.String(), info.ID)
		}
	}

	// Tags (zero-size components) are excluded from the blueprint's data layout.
	// This prevents the factory from attempting to allocate memory or return pointers for non-existent columns.
	if info.Size == 0 {
		return fmt.Errorf("cannot add %s: tags are not allowed in data-driven blueprints", info.Type.String())
	}

	b.compInfos = append(b.compInfos, info)
	return nil
}

// ExcludeType adds a component ID to the exclusion list.
// Returns an error if the ID is already marked for exclusion.
func (b *Blueprint) ExcludeType(id ComponentID) error {
	for _, existingID := range b.exCompIDs {
		if existingID == id {
			return fmt.Errorf("component ID %d is already in the exclusion list", id)
		}
	}
	b.exCompIDs = append(b.exCompIDs, id)
	return nil
}
