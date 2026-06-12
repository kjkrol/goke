package arch

import "github.com/kjkrol/goke/internal/core"

type ArchetypeGraph struct {
	edgesNext [core.MaxComponents]core.ArchetypeId
	edgesPrev [core.MaxComponents]core.ArchetypeId
}

func (a *ArchetypeGraph) Reset() {
	clear(a.edgesNext[:])
	clear(a.edgesPrev[:])
}

func (a *ArchetypeGraph) CountNextEdges() int {
	return countNonZeros(a.edgesNext)
}

func (a *ArchetypeGraph) CountPrevEdges() int {
	return countNonZeros(a.edgesPrev)
}

func countNonZeros(edges [core.MaxComponents]core.ArchetypeId) int {
	count := 0
	for _, edge := range edges {
		if edge != 0 {
			count++
		}
	}
	return count
}
