package arch

import (
	"fmt"
	"hash/maphash"
	"strings"
)

// MutCheck tells, from one Check to the next, a chunk whose bytes changed while its change
// counter and its table's version did not: a write that went past the counter.
type MutCheck struct {
	seed   maphash.Seed
	tables []tableSums // by archetype
}

type tableSums struct {
	ver    uint32
	seen   bool
	chunks []chunkSum // by chunk index
}

type chunkSum struct {
	ptr uintptr
	ver uint32
	sum uint64
}

// NewMutCheck is a check that has seen no chunk yet.
func NewMutCheck() *MutCheck { return &MutCheck{seed: maphash.MakeSeed()} }

// Check panics, naming the archetype, on the first chunk of c whose bytes changed since the
// last Check while neither its counter nor its table's version moved; then it remembers c.
func (m *MutCheck) Check(c *Catalog) {
	if n := int(c.Len()); len(m.tables) < n {
		m.tables = append(m.tables, make([]tableSums, n-len(m.tables))...)
	}
	for id := RootID; id < c.Len(); id++ {
		m.checkTable(&c.Archetypes[id], &m.tables[id])
	}
}

// Reset forgets every chunk seen.
func (m *MutCheck) Reset() { m.tables = nil }

func (m *MutCheck) checkTable(a *Archetype, s *tableSums) {
	t := &a.Table
	same := s.seen && s.ver == t.Version()
	prev := s.chunks
	s.chunks = s.chunks[:0]
	for i := range t.NumChunks() {
		ptr := t.ChunkPtrAt(i)
		now := chunkSum{
			ptr: uintptr(ptr),
			ver: *t.VerAt(ptr),
			sum: maphash.Bytes(m.seed, t.ChunkBytes(i)),
		}
		if same && int(i) < len(prev) {
			if was := prev[i]; was.ptr == now.ptr && was.ver == now.ver && was.sum != now.sum {
				panic(fmt.Sprintf("goke: a chunk of archetype %s changed since the last Sync "+
					"without its counter: a write through Comp where Mut belongs", name(a)))
			}
		}
		s.chunks = append(s.chunks, now)
	}
	s.ver, s.seen = t.Version(), true
}

func name(a *Archetype) string {
	defs := a.Composition().Defs
	names := make([]string, len(defs))
	for i, d := range defs {
		names[i] = fmt.Sprint(d.Type)
	}
	return "[" + strings.Join(names, " ") + "]"
}
