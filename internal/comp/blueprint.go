package comp

import (
	"fmt"
	"slices"
)

type Blueprint struct {
	CompInfos []Meta
	TagIDs    []ID
	ExCompIDs []ID
}

// Init applies opts against mi, populating b in place.
// Panics if any opt returns an error.
func (b *Blueprint) Init(mi *MetaIndex, opts ...BlueprintOpt) {
	for _, opt := range opts {
		if err := opt(b, mi); err != nil {
			panic(fmt.Sprintf("comp: blueprint option: %v", err))
		}
	}
}

// Compose derives a Composition from the Blueprint without requiring a MetaIndex.
func (b *Blueprint) Compose() Composition {
	return Composition{Mask: NewMask(b), Metas: b.CompInfos}
}

func (b *Blueprint) Comp(meta Meta) error {
	for _, existing := range b.CompInfos {
		if existing.ID == meta.ID {
			return fmt.Errorf("component %s (ID: %d) is already defined in this blueprint", meta.Type.String(), meta.ID)
		}
	}
	if meta.Size == 0 {
		return fmt.Errorf("cannot add %s: tags are not allowed in data-driven blueprints", meta.Type.String())
	}
	b.CompInfos = append(b.CompInfos, meta)
	return nil
}

func (b *Blueprint) Tag(tagId ID) error {
	if slices.Contains(b.TagIDs, tagId) {
		return fmt.Errorf("tag with ID %d is already defined in this blueprint", tagId)
	}
	b.TagIDs = append(b.TagIDs, tagId)
	return nil
}

func (b *Blueprint) Exclude(id ID) error {
	if slices.Contains(b.ExCompIDs, id) {
		return fmt.Errorf("component ID %d is already in the exclusion list", id)
	}
	b.ExCompIDs = append(b.ExCompIDs, id)
	return nil
}
