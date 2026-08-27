package query

import (
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/colstore"
	"github.com/kjkrol/goke/v3/iter"
)

type BakedTable struct {
	ArchID      arch.ID
	Table       *colstore.Table
	CompOffsets []uintptr
	OptOffsets  []uintptr
	OptPresent  []bool
}

func (bt *BakedTable) FillCursorNext(cur *iter.Cursor, from int) (int, bool) {
	idx, ok := bt.Table.FillCursorNext(cur, from, bt.CompOffsets)
	cur.OptOffsets = bt.OptOffsets
	cur.OptPresent = bt.OptPresent
	return idx, ok
}
