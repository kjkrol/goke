package colstore

import "github.com/kjkrol/goke/internal/comp"

type columnPos uint8

const (
	entityColumnPos    = columnPos(0)
	firstDataColumnPos = columnPos(1)
	invalidColumnPos   = columnPos(comp.MaxComponents + 1)
)

// columnIndex is a fixed-size array mapping each comp.ID to a columnPos —
// the local index of that component's Column within Table.columns.
// Lookups are a single array read: O(1) with no hashing or scanning.
type columnIndex [comp.MaxComponents]columnPos

func (m *columnIndex) Reset() {
	for i := range m {
		m[i] = invalidColumnPos
	}
}

func (m *columnIndex) Set(globalID comp.ID, localIdx columnPos) {
	m[globalID] = localIdx
}

func (m *columnIndex) Get(globalID comp.ID) columnPos {
	return m[globalID]
}
