package arch

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/soa"
)

type EntityIndex struct {
	entries []EntityLocation
}

func NewEntityIndex(initialCap int) EntityIndex {
	return EntityIndex{
		entries: make([]EntityLocation, initialCap),
	}
}

func (s *EntityIndex) Reset() {
	clear(s.entries)
}

func (s *EntityIndex) Get(entityID uid.UID64) (EntityLocation, bool) {
	index, gen := entityID.Unpack()
	if index >= uint32(len(s.entries)) {
		return EntityLocation{}, false
	}

	loc := s.entries[index]

	if loc.ArchId == NullID || loc.Generation != gen {
		return EntityLocation{}, false
	}

	return loc, true
}

func (s *EntityIndex) Upsert(entityID uid.UID64, archId ID, pos soa.BlockPos) {
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

func (s *EntityIndex) Clear(entityID uid.UID64) {
	index, gen := entityID.Unpack()
	if index >= uint32(cap(s.entries)) {
		return
	}

	if s.entries[index].Generation == gen {
		s.entries[index] = EntityLocation{ArchId: NullID}
	}
}

func (s *EntityIndex) grow(minLen uint32) {
	newCap := max(uint32(len(s.entries))*2, minLen)
	newEntries := make([]EntityLocation, newCap)
	copy(newEntries, s.entries)
	s.entries = newEntries
}
