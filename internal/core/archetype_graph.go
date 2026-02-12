package core

type ArchetypeGraph struct {
	edgesNext [MaxComponents]ArchetypeId
	edgesPrev [MaxComponents]ArchetypeId
}

// CountNextEdges remains as is (or use a stored counter if needed)
func (a *ArchetypeGraph) CountNextEdges() int {
	return countNonZeros(a.edgesNext)
}

func (a *ArchetypeGraph) CountPrevEdges() int {
	return countNonZeros(a.edgesPrev)
}

func countNonZeros(edges [MaxComponents]ArchetypeId) int {
	count := 0
	for _, edge := range edges {
		if edge != 0 {
			count++
		}
	}
	return count
}
