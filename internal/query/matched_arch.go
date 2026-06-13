package query

import "github.com/kjkrol/goke/internal/colstore"

type MatchedArch struct {
	Table            *colstore.Table
	EntityPageOffset uintptr
	CompOffsets      []uintptr
	CompSizes        []uintptr
}
