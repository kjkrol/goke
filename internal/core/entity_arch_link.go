package core

import "github.com/kjkrol/uid"

type EntityArchLink struct {
	ArchId     ArchetypeId
	PageIdx    PageIdx // Index of the memory page (Page) in Archetype.Memory.Pages
	PageRow    PageRow // Index of the row within that specific Page
	Generation uint32  // Entity generation for validation
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

// Get returns the link only if the generation matches.
func (s *EntityLinkStore) Get(entity uid.UID64) (EntityArchLink, bool) {
	index, gen := entity.Unpack()
	if index >= uint32(len(s.links)) {
		return EntityArchLink{}, false
	}

	link := s.links[index]

	// Combined check: empty slot OR stale generation.
	if link.ArchId == NullArchetypeId || link.Generation != gen {
		return EntityArchLink{}, false
	}

	return link, true
}

// Update updates the entity's location using the new Page Index and Page Row.
func (s *EntityLinkStore) Update(entity uid.UID64, archId ArchetypeId, pageIdx PageIdx, row PageRow) {
	index, gen := entity.Unpack()
	if index >= uint32(cap(s.links)) {
		s.grow(index + 1)
	}

	// Store location (Page Index + Row) AND the current generation
	s.links[index] = EntityArchLink{
		ArchId:     archId,
		PageIdx:    pageIdx,
		PageRow:    row,
		Generation: gen,
	}
}

func (s *EntityLinkStore) Clear(entity uid.UID64) {
	index, gen := entity.Unpack()
	if index >= uint32(cap(s.links)) {
		return
	}

	// Double-check: only clear if generations match!
	// If they don't, it means this entity is already gone or replaced.
	if s.links[index].Generation == gen {
		s.links[index] = EntityArchLink{
			ArchId:     NullArchetypeId,
			PageIdx:    0,
			PageRow:    0,
			Generation: 0, // Reset to 0 to prevent any stale handle matches
		}
	}
}

func (s *EntityLinkStore) grow(minLen uint32) {
	newCap := max(uint32(len(s.links))*2, minLen)

	newLinks := make([]EntityArchLink, newCap)
	copy(newLinks, s.links)
	s.links = newLinks
}
