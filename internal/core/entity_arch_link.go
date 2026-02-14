package core

type EntityArchLink struct {
	ArchId     ArchetypeId
	ChunkIdx   ChunkIdx // Index of the memory page (Chunk) in Archetype.Memory.Pages
	ChunkRow   ChunkRow // Index of the row within that specific Chunk
	Generation uint32   // Entity generation for validation
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
func (s *EntityLinkStore) Get(entity Entity) (EntityArchLink, bool) {
	index := entity.Index()
	if index >= uint32(cap(s.links)) {
		return EntityArchLink{}, false
	}

	link := s.links[index]

	// Safety check: compare generations.
	// We also check ArchId to ensure the link is not empty/invalid.
	if link.ArchId == NullArchetypeId || link.Generation != entity.Generation() {
		return EntityArchLink{}, false
	}

	return link, true
}

// Update updates the entity's location using the new Chunk Index and Chunk Row.
func (s *EntityLinkStore) Update(entity Entity, archId ArchetypeId, chunkIdx ChunkIdx, row ChunkRow) {
	index := entity.Index()
	if index >= uint32(cap(s.links)) {
		s.grow(index + 1)
	}

	// Store location (Chunk Index + Row) AND the current generation
	s.links[index] = EntityArchLink{
		ArchId:     archId,
		ChunkIdx:   chunkIdx,
		ChunkRow:   row,
		Generation: entity.Generation(),
	}
}

func (s *EntityLinkStore) Clear(entity Entity) {
	index := entity.Index()
	if index >= uint32(cap(s.links)) {
		return
	}

	// Double-check: only clear if generations match!
	// If they don't, it means this entity is already gone or replaced.
	if s.links[index].Generation == entity.Generation() {
		s.links[index] = EntityArchLink{
			ArchId:     NullArchetypeId,
			ChunkIdx:   0,
			ChunkRow:   0,
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
