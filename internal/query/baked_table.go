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

// PointCursor moves cur to pos within this baked table (chunk base + slot).
// The caller is responsible for having set cur.Offsets to this table's
// CompOffsets beforehand.
func (bt *BakedTable) PointCursor(cur *iter.Cursor, pos colstore.Pos) {
	bt.Table.PointCursor(cur, pos)
}
