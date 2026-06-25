package addr

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/colstore"
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

// GetUnchecked returns the Entry for the given entity ID without validating
// that it is alive (no bounds check, no generation check). The caller must
// already know the entity is alive — e.g. it was returned by a prior Get on
// this Index. Misuse silently returns a stale or zero Entry.
func (s *Index) GetUnchecked(entityID uid.UID64) Entry {
	index, _ := entityID.Unpack()
	return s.entries[index]
}

// Get returns the Entry for the given entity ID, or false if the ID is invalid
// (wrong generation) or has no registered address.
func (s *Index) Get(entityID uid.UID64) (Entry, bool) {
	index, gen := entityID.Unpack()
	// Unsigned compare (matching the index's own width) lets the compiler
	// prove the subsequent s.entries[index] access in-bounds and elide its
	// runtime bounds check.
	if uint(index) >= uint(len(s.entries)) {
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
	// len, not cap: s.entries is only ever replaced via grow's make([]Entry,
	// newCap) (never re-sliced), so len == cap always — but the BCE prover
	// reasons about len for the write below, not cap.
	if uint(index) >= uint(len(s.entries)) {
		s.grow(index + 1)
	}
	s.entries[index] = Entry{ArchId: archId, Pos: pos, Gen: gen}
}

func (s *Index) Clear(entityID uid.UID64) {
	index, gen := entityID.Unpack()
	if uint(index) >= uint(len(s.entries)) {
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
