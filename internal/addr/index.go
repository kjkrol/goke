package addr

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
)

// Index is a flat slice keyed by the numeric index extracted from a [uid.UID64].
// It resolves an entity ID to its [Entry] in O(1) — no hash map, no scanning.
type Index struct {
	entries []Entry
}

func (s *Index) Init(initialCap int) {
	s.entries = make([]Entry, initialCap)
}

func (s *Index) Reset() {
	clear(s.entries)
}

// Get returns the Entry for the given entity ID, or false if the ID is invalid
// (wrong generation) or has no registered address.
func (s *Index) Get(entityID uid.UID64) (Entry, bool) {
	index, gen := entityID.Unpack()
	if index >= uint32(len(s.entries)) {
		return Entry{}, false
	}
	e := s.entries[index]
	if e.ArchId == arch.NullID || e.Gen != gen {
		return Entry{}, false
	}
	return e, true
}

func (s *Index) Upsert(entityID uid.UID64, archId arch.ID, pos colstore.Pos) {
	index, gen := entityID.Unpack()
	if index >= uint32(cap(s.entries)) {
		s.grow(index + 1)
	}
	s.entries[index] = Entry{ArchId: archId, Pos: pos, Gen: gen}
}

func (s *Index) Clear(entityID uid.UID64) {
	index, gen := entityID.Unpack()
	if index >= uint32(cap(s.entries)) {
		return
	}
	if s.entries[index].Gen == gen {
		s.entries[index] = Entry{ArchId: arch.NullID}
	}
}

func (s *Index) EnsureCap(minLen uint32) {
	if uint32(cap(s.entries)) < minLen {
		s.grow(minLen)
	}
}

func (s *Index) UpsertUnchecked(entityID uid.UID64, archId arch.ID, pos colstore.Pos) {
	index, gen := entityID.Unpack()
	s.entries[index] = Entry{ArchId: archId, Pos: pos, Gen: gen}
}

func (s *Index) grow(minLen uint32) {
	newCap := max(uint32(len(s.entries))*2, minLen)
	newEntries := make([]Entry, newCap)
	copy(newEntries, s.entries)
	s.entries = newEntries
}
