package query

import (
	"github.com/kjkrol/goke/internal/colstore"
	"github.com/kjkrol/goke/iter"
)

type BakedTable struct {
	Table       *colstore.Table
	CompOffsets []uintptr
}

func (bt *BakedTable) FillCursorNext(cur *iter.Cursor, from int) (int, bool) {
	return bt.Table.FillCursorNext(cur, from, bt.CompOffsets)
}

func (bt *BakedTable) FillCursorAt(cur *iter.Cursor, pos colstore.Pos) {
	bt.Table.FillCursorAt(cur, pos, bt.CompOffsets)
}
