package arch

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/internal/mem"
)

type EntityArchLink struct {
	ArchId     core.ArchetypeId
	PageIdx    mem.PageIdx
	PageSlot   mem.PageSlot
	Generation uint32
}

type EntityLinkStore struct {
	links []EntityArchLink
}

func NewEntityLinkStore(initialCap int) EntityLinkStore {
	return EntityLinkStore{
		links: make([]EntityArchLink, initialCap),
	}
}

func (s *EntityLinkStore) Reset() {
	clear(s.links)
}

func (s *EntityLinkStore) Get(entity uid.UID64) (EntityArchLink, bool) {
	index, gen := entity.Unpack()
	if index >= uint32(len(s.links)) {
		return EntityArchLink{}, false
	}

	link := s.links[index]

	if link.ArchId == core.NullArchetypeId || link.Generation != gen {
		return EntityArchLink{}, false
	}

	return link, true
}

func (s *EntityLinkStore) Update(entity uid.UID64, archId core.ArchetypeId, pageIdx mem.PageIdx, slot mem.PageSlot) {
	index, gen := entity.Unpack()
	if index >= uint32(cap(s.links)) {
		s.grow(index + 1)
	}

	s.links[index] = EntityArchLink{
		ArchId:     archId,
		PageIdx:    pageIdx,
		PageSlot:   slot,
		Generation: gen,
	}
}

func (s *EntityLinkStore) Clear(entity uid.UID64) {
	index, gen := entity.Unpack()
	if index >= uint32(cap(s.links)) {
		return
	}

	if s.links[index].Generation == gen {
		s.links[index] = EntityArchLink{
			ArchId:     core.NullArchetypeId,
			PageIdx:    0,
			PageSlot:   0,
			Generation: 0,
		}
	}
}

func (s *EntityLinkStore) grow(minLen uint32) {
	newCap := max(uint32(len(s.links))*2, minLen)
	newLinks := make([]EntityArchLink, newCap)
	copy(newLinks, s.links)
	s.links = newLinks
}
