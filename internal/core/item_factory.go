package core

type ItemFactory struct {
	Reg       *Registry
	Mask      ArchetypeMask
	CompInfos []ComponentInfo
	ArchId    ArchetypeId
}

func NewItemFactory(blueprint *Blueprint) *ItemFactory {
	var mask ArchetypeMask

	for _, info := range blueprint.compInfos {
		mask = mask.Set(info.ID)
	}

	for _, tag := range blueprint.tagIDs {
		mask = mask.Set(tag)
	}

	archId := blueprint.Reg.ArchetypeRegistry.getOrRegister(mask)

	return &ItemFactory{
		Reg:       blueprint.Reg,
		Mask:      mask,
		CompInfos: blueprint.compInfos,
		ArchId:    archId,
	}
}
