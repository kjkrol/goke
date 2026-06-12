package reg

import "github.com/kjkrol/goke/internal/core"

type ItemFactory struct {
	Reg       *Registry
	Mask      core.ArchetypeMask
	CompInfos []core.ComponentInfo
	ArchId    core.ArchetypeId
}

func NewItemFactory(blueprint *Blueprint) *ItemFactory {
	var mask core.ArchetypeMask

	for _, info := range blueprint.compInfos {
		mask = mask.Set(info.ID)
	}

	for _, tag := range blueprint.tagIDs {
		mask = mask.Set(tag)
	}

	archId := blueprint.Reg.ArchetypeRegistry.GetOrRegister(mask)

	return &ItemFactory{
		Reg:       blueprint.Reg,
		Mask:      mask,
		CompInfos: blueprint.compInfos,
		ArchId:    archId,
	}
}
