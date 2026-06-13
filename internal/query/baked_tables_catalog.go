package query

import (
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
)

type BakedTablesCatalog struct {
	BakedTables    []BakedTable
	archTableIndex []int32
}

// Add bakes the archetype into a BakedTable and registers it in the catalog.
// metas defines which component columns are precomputed for iteration.
func (c *BakedTablesCatalog) Add(archetype *arch.Archetype, metas []comp.Meta) {
	offsets := make([]uintptr, len(metas))
	sizes := make([]uintptr, len(metas))
	for idx, meta := range metas {
		col := archetype.Table.GetColumn(meta.ID)
		if col == nil {
			continue
		}
		offsets[idx] = col.Offset
		sizes[idx] = col.CompSize
	}

	c.BakedTables = append(c.BakedTables, BakedTable{
		Table:       &archetype.Table,
		CompOffsets: offsets,
		CompSizes:   sizes,
	})

	if int(archetype.Id) >= len(c.archTableIndex) {
		c.grow(int(archetype.Id) + 1)
	}
	c.archTableIndex[archetype.Id] = int32(len(c.BakedTables) - 1)
}

// Get returns the BakedTable for the given archetype ID, or nil if not matched.
func (c *BakedTablesCatalog) Get(archID arch.ID) *BakedTable {
	if int(archID) >= len(c.archTableIndex) {
		return nil
	}
	idx := c.archTableIndex[archID]
	if idx == -1 {
		return nil
	}
	return &c.BakedTables[idx]
}

func (c *BakedTablesCatalog) Clear() {
	clear(c.BakedTables)
	c.BakedTables = nil
	clear(c.archTableIndex)
	c.archTableIndex = nil
}

func (c *BakedTablesCatalog) grow(minLen int) {
	newCap := max(cap(c.archTableIndex)*2, minLen)
	grown := make([]int32, minLen, newCap)
	copy(grown, c.archTableIndex)
	for i := len(c.archTableIndex); i < minLen; i++ {
		grown[i] = -1
	}
	c.archTableIndex = grown
}
