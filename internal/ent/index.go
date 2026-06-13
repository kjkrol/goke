package ent

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/soa"
)

type Index struct {
	entries []EntityLocation
}

func (s *Index) Init(initialCap int) {
	s.entries = make([]EntityLocation, initialCap)
}

func (s *Index) Reset() {
	clear(s.entries)
}

func (s *Index) Get(entityID uid.UID64) (EntityLocation, bool) {
	index, gen := entityID.Unpack()
	if index >= uint32(len(s.entries)) {
		return EntityLocation{}, false
	}

	loc := s.entries[index]

	if loc.ArchId == arch.NullID || loc.Generation != gen {
		return EntityLocation{}, false
	}

	return loc, true
}

func (s *Index) Upsert(entityID uid.UID64, archId arch.ID, pos soa.BlockPos) {
	index, gen := entityID.Unpack()
	if index >= uint32(cap(s.entries)) {
		s.grow(index + 1)
	}

	s.entries[index] = EntityLocation{
		ArchId:     archId,
		Pos:        pos,
		Generation: gen,
	}
}

func (s *Index) Clear(entityID uid.UID64) {
	index, gen := entityID.Unpack()
	if index >= uint32(cap(s.entries)) {
		return
	}

	if s.entries[index].Generation == gen {
		s.entries[index] = EntityLocation{ArchId: arch.NullID}
	}
}

func (s *Index) grow(minLen uint32) {
	newCap := max(uint32(len(s.entries))*2, minLen)
	newEntries := make([]EntityLocation, newCap)
	copy(newEntries, s.entries)
	s.entries = newEntries
}
