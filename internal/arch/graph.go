package arch

import "github.com/kjkrol/goke/v2/internal/comp"

type Graph struct {
	edgesNext [comp.MaxComponents]ID
	edgesPrev [comp.MaxComponents]ID
}

func (a *Graph) Reset() {
	clear(a.edgesNext[:])
	clear(a.edgesPrev[:])
}

func (a *Graph) CountNextEdges() int {
	return countNonZeros(a.edgesNext)
}

func (a *Graph) CountPrevEdges() int {
	return countNonZeros(a.edgesPrev)
}

func (g *Graph) linkNext(other *Graph, selfID ID, otherID ID, compID comp.ID) {
	g.edgesNext[compID] = otherID
	other.edgesPrev[compID] = selfID
}

func (g *Graph) linkPrev(other *Graph, selfID ID, otherID ID, compID comp.ID) {
	g.edgesPrev[compID] = otherID
	other.edgesNext[compID] = selfID
}

func countNonZeros(edges [comp.MaxComponents]ID) int {
	count := 0
	for _, edge := range edges {
		if edge != 0 {
			count++
		}
	}
	return count
}
