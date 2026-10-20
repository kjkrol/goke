package query

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
)

// every way of standing on an entity points Cursor.Ver() at the counter of the entity's own
// chunk, and a walk's snapshot carries the counter's value
func TestMatcher_CursorVerIsTheChunksCounter(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	var pos iter.ArrayRef[iterPos]
	posOpt := comp.Track(&pos)
	var spec comp.AccessSpec
	spec.Init(cc, posOpt)
	f := em.CreateFactory(spec)
	var ids []uid.UID64
	f.Create(3000)
	for f.Next() {
		ids = append(ids, f.IDs...)
	}
	m := NewMatcher(cat, posOpt)

	want := func(e uid.UID64) *uint32 {
		entry, _ := em.AddressBook.Get(e)
		return em.ArchCatalog.Archetypes[entry.ArchID].Table.VerAt(entry.ChunkPtr)
	}
	first, last := ids[0], ids[len(ids)-1]
	if !m.Seek(last) || m.Cursor.Ver() != want(last) {
		t.Error("Seek does not point Ver at the entity's chunk")
	}
	if !m.SeekH(first) || m.Cursor.Ver() != want(first) {
		t.Error("SeekH does not point Ver at the entity's chunk")
	}
	for m.Pick([]uid.UID64{first, last}); m.Next(); {
		if m.Cursor.Ver() != want(m.Entity) {
			t.Errorf("Pick does not point Ver at the chunk of %v", m.Entity)
		}
	}
	chunks := 0
	for m.All(); m.Next(); chunks++ {
		*m.Cursor.Ver() += uint32(chunks)
		if s := m.ChunkSnapshot(); s.Ver != *m.Cursor.Ver() || m.Cursor.Ver() != want(m.Cursor.IDs[0]) {
			t.Errorf("chunk %d: snapshot Ver %d, counter %d", chunks, s.Ver, *m.Cursor.Ver())
		}
	}
	if chunks < 2 {
		t.Fatalf("the walk met %d chunks; the test needs several", chunks)
	}
}
