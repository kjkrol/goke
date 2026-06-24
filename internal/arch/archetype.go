package arch

import (
	"github.com/kjkrol/goke/v2/internal/colstore"
	"github.com/kjkrol/goke/v2/internal/comp"
)

type ID uint16

const (
	NullID = ID(0)
	RootID = ID(1)
	MaxID  = ID(4096)
)

type Archetype struct {
	Id    ID
	Table colstore.Table
	set   comp.Composition
	graph *Graph
}

func (a *Archetype) Mask() comp.Mask { return a.set.Mask }

func (a *Archetype) Reset() {
	a.Table.Clear()
	if a.graph != nil {
		a.graph.Reset()
	}
	a.Id = NullID
	a.set = comp.Composition{}
}

func (a *Archetype) Init(archId ID, set comp.Composition) {
	a.Id = archId
	a.set = set
	a.graph = &Graph{}
	a.Table.Init(set.Defs)
}

func (a *Archetype) Len() int {
	return int(a.Table.Len())
}
