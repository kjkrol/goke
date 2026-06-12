package mem

import (
	"unsafe"

	"github.com/kjkrol/goke/internal/core"
)

type LocalColumnID uint8

const (
	EntityColumnIndex    = LocalColumnID(0)
	FirstDataColumnIndex = LocalColumnID(1)
	InvalidLocalID       = LocalColumnID(core.MaxComponents + 1)
)

type ColumnMap [core.MaxComponents]LocalColumnID

func (m *ColumnMap) Reset() {
	for i := range m {
		m[i] = InvalidLocalID
	}
}

func (m *ColumnMap) Set(globalID core.ComponentID, localIdx LocalColumnID) {
	m[globalID] = localIdx
}

func (m *ColumnMap) Get(globalID core.ComponentID) LocalColumnID {
	return m[globalID]
}

type Column struct {
	CompID     core.ComponentID
	ItemSize   uintptr
	PageOffset uintptr
}

func (c *Column) GetPointer(page *Page, pageSlot PageSlot) unsafe.Pointer {
	return unsafe.Add(page.Ptr, c.PageOffset+uintptr(pageSlot)*c.ItemSize)
}

func (c *Column) GetColumnStart(page *Page) unsafe.Pointer {
	return unsafe.Add(page.Ptr, c.PageOffset)
}
