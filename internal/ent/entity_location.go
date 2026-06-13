package ent

import (
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/soa"
)

type EntityLocation struct {
	ArchId     arch.ID
	Pos        soa.BlockPos
	Generation uint32
}
