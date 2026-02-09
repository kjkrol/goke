package core

type EntityArchLink struct {
	ArchId     ArchetypeId
	Row        ArchRow
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

// Get returns the link only if the generation matches.
func (s *EntityLinkStore) Get(entity Entity) (EntityArchLink, bool) {
	index := entity.Index()
	if index >= uint32(len(s.links)) {
		return EntityArchLink{}, false
	}

	link := s.links[index]

	// Safety check: compare generations
	if link.ArchId == NullArchetypeId || link.Generation != entity.Generation() {
		return EntityArchLink{}, false
	}

	return link, true
}

func (s *EntityLinkStore) Update(entity Entity, archId ArchetypeId, row ArchRow) {
	index := entity.Index()
	if index >= uint32(len(s.links)) {
		s.grow(index + 1)
	}
	// Store both the location AND the current generation
	s.links[index] = EntityArchLink{
		ArchId:     archId,
		Row:        row,
		Generation: entity.Generation(),
	}
}

func (s *EntityLinkStore) Clear(entity Entity) {
	index := entity.Index()
	if index >= uint32(len(s.links)) {
		return
	}

	// Double-check: only clear if generations match!
	// If they don't, it means this entity is already gone or replaced.
	if s.links[index].Generation == entity.Generation() {
		s.links[index] = EntityArchLink{
			ArchId:     NullArchetypeId,
			Row:        0,
			Generation: 0, // Reset to 0 to prevent any stale handle matches
		}
	}
}

func (s *EntityLinkStore) grow(minLen uint32) {
	newCap := uint32(len(s.links)) * 2
	if newCap < minLen {
		newCap = minLen
	}
	newLinks := make([]EntityArchLink, newCap)
	copy(newLinks, s.links)
	s.links = newLinks
}
