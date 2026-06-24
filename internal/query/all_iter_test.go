package query

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/iter"
	"github.com/kjkrol/uid"
)

// iterPos and iterVel are shared by all_iter_test, filter_iter_test, and view_test.
type (
	iterPos struct{ X, Y float32 }
	iterVel struct{ VX, VY float32 }
)

func TestAllIter_EmptyMatcher(t *testing.T) {
	cat, _, _ := newQueryCatalog()
	m := NewMatcher(cat)

	m.All()
	if m.Next() {
		t.Error("Next() on empty matcher should return false")
	}
	if m.Cursor.IDs != nil {
		t.Error("Chunk should be nil after exhaustion")
	}
}

func TestAllIter_SingleEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	posOpt := comp.Track(&pos)
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, posOpt)
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 1, Y: 2}

	m := NewMatcher(cat, posOpt)

	var visited []uid.UID64
	m.All()
	for m.Next() {
		posSlice := pos.Slice(&m.Cursor)
		for i, entity := range m.Cursor.IDs {
			_ = posSlice[i]
			visited = append(visited, entity)
		}
	}

	if len(visited) != 1 || visited[0] != e {
		t.Errorf("expected [%v], got %v", e, visited)
	}
}

func TestAllIter_SliceValues(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	posOpt := comp.Track(&pos)
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, posOpt)
	f := em.CreateFactory(accessSpec)

	f.Create(1)
	f.Next()
	e1 := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 10, Y: 20}

	f.Create(1)
	f.Next()
	e2 := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 30, Y: 40}

	m := NewMatcher(cat, posOpt)

	got := map[uid.UID64]iterPos{}
	m.All()
	for m.Next() {
		posSlice := pos.Slice(&m.Cursor)
		for i, entity := range m.Cursor.IDs {
			got[entity] = posSlice[i]
		}
	}

	if got[e1] != (iterPos{X: 10, Y: 20}) {
		t.Errorf("e1: want {10,20} got %v", got[e1])
	}
	if got[e2] != (iterPos{X: 30, Y: 40}) {
		t.Errorf("e2: want {30,40} got %v", got[e2])
	}
}

func TestAllIter_SliceInPlaceMutation(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(&pos))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	*pos.At(&f.Cursor) = iterPos{X: 5, Y: 3}

	m := NewMatcher(cat, comp.Track(&pos))

	m.All()
	for m.Next() {
		posSlice := pos.Slice(&m.Cursor)
		for i := range m.Cursor.IDs {
			posSlice[i].X += posSlice[i].Y
		}
	}

	// re-read from live memory via a second pass
	m.All()
	for m.Next() {
		posSlice := pos.Slice(&m.Cursor)
		for range m.Cursor.IDs {
			if posSlice[0].X != 8 {
				t.Errorf("expected X=8 after mutation, got %v", posSlice[0].X)
			}
		}
	}
}

func TestAllIter_MultipleArchetypes(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	// archetype A: pos only
	var posA iter.ArrayRef[iterPos]
	var accessSpecA comp.AccessSpec
	accessSpecA.Init(cc, comp.Track(&posA))
	fA := em.CreateFactory(accessSpecA)
	fA.Create(1)
	fA.Next()
	eA := fA.IDs[0]
	*posA.At(&fA.Cursor) = iterPos{X: 1}

	// archetype B: pos + vel
	var posB iter.ArrayRef[iterPos]
	var accessSpecB comp.AccessSpec
	accessSpecB.Init(cc, comp.Track(&posB), comp.Track(new(iter.ArrayRef[iterVel])))
	fB := em.CreateFactory(accessSpecB)
	fB.Create(1)
	fB.Next()
	eB := fB.IDs[0]
	*posB.At(&fB.Cursor) = iterPos{X: 2}

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
	m := NewMatcher(cat, _trackOpt0)

	visited := map[uid.UID64]bool{}
	m.All()
	for m.Next() {
		for _, e := range m.Cursor.IDs {
			visited[e] = true
		}
	}

	if !visited[eA] || !visited[eB] {
		t.Errorf("expected both entities visited, got %v", visited)
	}
}

func TestAllIter_ResetOnSecondCall(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	m := NewMatcher(cat, comp.Track(new(iter.ArrayRef[iterPos])))

	count := func() int {
		n := 0
		m.All()
		for m.Next() {
			n += len(m.Cursor.IDs)
		}
		return n
	}

	if n := count(); n != 1 {
		t.Errorf("first All() call: expected 1 entity, got %d", n)
	}
	if n := count(); n != 1 {
		t.Errorf("second All() call: expected 1 entity, got %d", n)
	}
}
