package reg

import (
	"fmt"
	"slices"

	"github.com/kjkrol/goke/internal/core"
)

type Blueprint struct {
	Reg       *Registry
	compInfos []core.ComponentInfo
	tagIDs    []core.ComponentID
	exCompIDs []core.ComponentID
}

func NewBlueprint(reg *Registry) *Blueprint {
	return &Blueprint{
		Reg:       reg,
		compInfos: make([]core.ComponentInfo, 0, 8),
		tagIDs:    make([]core.ComponentID, 0, 8),
		exCompIDs: make([]core.ComponentID, 0, 8),
	}
}

func (b *Blueprint) WithTag(tagId core.ComponentID) error {
	if slices.Contains(b.tagIDs, tagId) {
		return fmt.Errorf("tag with ID %d is already defined in this blueprint", tagId)
	}
	b.tagIDs = append(b.tagIDs, tagId)
	return nil
}

func (b *Blueprint) WithComp(info core.ComponentInfo) error {
	for _, existing := range b.compInfos {
		if existing.ID == info.ID {
			return fmt.Errorf("component %s (ID: %d) is already defined in this blueprint", info.Type.String(), info.ID)
		}
	}

	if info.Size == 0 {
		return fmt.Errorf("cannot add %s: tags are not allowed in data-driven blueprints", info.Type.String())
	}

	b.compInfos = append(b.compInfos, info)
	return nil
}

func (b *Blueprint) ExcludeType(id core.ComponentID) error {
	if slices.Contains(b.exCompIDs, id) {
		return fmt.Errorf("component ID %d is already in the exclusion list", id)
	}
	b.exCompIDs = append(b.exCompIDs, id)
	return nil
}
