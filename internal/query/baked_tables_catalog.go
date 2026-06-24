package query

import (
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/comp"
)

type BakedTablesCatalog struct {
	BakedTables    []BakedTable
	archTableIndex []int32
}

// Add bakes the archetype into a BakedTable and registers it in the catalog.
// compIDs defines which component columns are precomputed for iteration.
func (c *BakedTablesCatalog) Add(archetype *arch.Archetype, compIDs []comp.ID) {
	c.BakedTables = append(c.BakedTables, BakedTable{
		Table:       &archetype.Table,
		CompOffsets: archetype.Table.BakeOffsets(compIDs),
	})

	if int(archetype.Id) >= len(c.archTableIndex) {
		c.grow(int(archetype.Id) + 1)
	}
	c.archTableIndex[archetype.Id] = int32(len(c.BakedTables) - 1)
}

// Get returns the BakedTable for the given archetype ID, or nil if not matched.
func (c *BakedTablesCatalog) Get(archID arch.ID) *BakedTable {
	// Unsigned compare enables bounds-check elimination on the index access;
	// the single trailing return nil covers both out-of-range and unmatched (-1).
	if uint(archID) < uint(len(c.archTableIndex)) {
		if idx := c.archTableIndex[archID]; idx >= 0 {
			return &c.BakedTables[idx]
		}
	}
	return nil
}

func (c *BakedTablesCatalog) Clear() {
	clear(c.BakedTables)
	c.BakedTables = nil
	clear(c.archTableIndex)
	c.archTableIndex = nil
}

func (c *BakedTablesCatalog) grow(minLen int) {
	oldLen := len(c.archTableIndex)
	if minLen <= cap(c.archTableIndex) {
		c.archTableIndex = c.archTableIndex[:minLen]
	} else {
		newCap := max(cap(c.archTableIndex)*2, minLen)
		grown := make([]int32, minLen, newCap)
		copy(grown, c.archTableIndex)
		c.archTableIndex = grown
	}
	for i := oldLen; i < minLen; i++ {
		c.archTableIndex[i] = -1
	}
}
