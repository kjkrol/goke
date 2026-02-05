package core

type Blueprint struct {
	Reg       *Registry
	compIDs   []ComponentID
	tagIDs    []ComponentID
	exCompIDs []ComponentID
}

func NewBlueprint(reg *Registry) *Blueprint {
	return &Blueprint{
		Reg:       reg,
		compIDs:   make([]ComponentID, 0, 8),
		tagIDs:    make([]ComponentID, 0, 8),
		exCompIDs: make([]ComponentID, 0, 8),
	}
}

func (b *Blueprint) WithTag(tagId ComponentID) {
	b.tagIDs = append(b.tagIDs, tagId)
}

func (b *Blueprint) WithComp(compId ComponentID) {
	b.compIDs = append(b.compIDs, compId)
}

func (b *Blueprint) ExcludeType(id ComponentID) {
	b.exCompIDs = append(b.exCompIDs, id)
}
