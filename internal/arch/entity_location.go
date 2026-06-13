package arch

import "github.com/kjkrol/goke/internal/soa"

type EntityLocation struct {
	ArchId     ID
	Pos        soa.BlockPos
	Generation uint32
}
