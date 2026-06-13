package colstore

import "github.com/kjkrol/goke/internal/comp"

type columnPos uint8

const (
	entityColumnPos    = columnPos(0)
	firstDataColumnPos = columnPos(1)
	invalidColumnPos   = columnPos(comp.MaxComponents + 1)
)

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
