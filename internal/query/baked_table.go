package query

import (
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/colstore"
	"github.com/kjkrol/goke/v2/iter"
)

type BakedTable struct {
	ArchID      arch.ID
	Table       *colstore.Table
	CompOffsets []uintptr
}

func (bt *BakedTable) FillCursorNext(cur *iter.Cursor, from int) (int, bool) {
	return bt.Table.FillCursorNext(cur, from, bt.CompOffsets)
}
