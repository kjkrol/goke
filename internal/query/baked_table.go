package query

import "github.com/kjkrol/goke/internal/colstore"

type BakedTable struct {
	Table       *colstore.Table
	CompOffsets []uintptr
	CompSizes   []uintptr
}
