package query

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
	"github.com/kjkrol/uid"
)

// iterPos and iterVel are shared by all_iter_test, filter_iter_test, and view_test.
type (
	iterPos struct{ X, Y float32 }
	iterVel struct{ VX, VY float32 }
)

func TestAllIter_EmptyView(t *testing.T) {
	cat, _, _ := newQueryCatalog()
	v := NewView(cat)

	v.All()
	if v.Next() {
		t.Error("Next() on empty view should return false")
	}
	if v.Cursor.IDs != nil {
		t.Error("Chunk should be nil after exhaustion")
	}
}

func TestAllIter_SingleEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.Col[iterPos]
	posOpt := comp.Track(&pos)
	var b comp.Blueprint
	b.Init(cc, posOpt)
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 1, Y: 2}

	v := NewView(cat, posOpt)

	var visited []uid.UID64
	v.All()
	for v.Next() {
		posSlice := pos.Slice(&v.Cursor)
		for i, entity := range v.Cursor.IDs {
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

	var pos iter.Col[iterPos]
	posOpt := comp.Track(&pos)
	var b comp.Blueprint
	b.Init(cc, posOpt)
	f := em.CreateFactory(b)

	f.Create(1)
	f.Next()
	e1 := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 10, Y: 20}

	f.Create(1)
	f.Next()
	e2 := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 30, Y: 40}

	v := NewView(cat, posOpt)

	got := map[uid.UID64]iterPos{}
	v.All()
	for v.Next() {
		posSlice := pos.Slice(&v.Cursor)
		for i, entity := range v.Cursor.IDs {
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

	var pos iter.Col[iterPos]
	var b comp.Blueprint
	b.Init(cc, comp.Track(&pos))
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	*pos.At(&f.Cursor) = iterPos{X: 5, Y: 3}

	v := NewView(cat, comp.Track(&pos))

	v.All()
	for v.Next() {
		posSlice := pos.Slice(&v.Cursor)
		for i := range v.Cursor.IDs {
			posSlice[i].X += posSlice[i].Y
		}
	}

	// re-read from live memory via a second pass
	v.All()
	for v.Next() {
		posSlice := pos.Slice(&v.Cursor)
		for range v.Cursor.IDs {
			if posSlice[0].X != 8 {
				t.Errorf("expected X=8 after mutation, got %v", posSlice[0].X)
			}
		}
	}
}

func TestAllIter_MultipleArchetypes(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	// archetype A: pos only
	var posA iter.Col[iterPos]
	var bA comp.Blueprint
	bA.Init(cc, comp.Track(&posA))
	fA := em.CreateFactory(bA)
	fA.Create(1)
	fA.Next()
	eA := fA.IDs[0]
	*posA.At(&fA.Cursor) = iterPos{X: 1}

	// archetype B: pos + vel
	var posB iter.Col[iterPos]
	var bB comp.Blueprint
	bB.Init(cc, comp.Track(&posB), comp.Track(new(iter.Col[iterVel])))
	fB := em.CreateFactory(bB)
	fB.Create(1)
	fB.Next()
	eB := fB.IDs[0]
	*posB.At(&fB.Cursor) = iterPos{X: 2}

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	v := NewView(cat, _trackOpt0)

	visited := map[uid.UID64]bool{}
	v.All()
	for v.Next() {
		for _, e := range v.Cursor.IDs {
			visited[e] = true
		}
	}

	if !visited[eA] || !visited[eB] {
		t.Errorf("expected both entities visited, got %v", visited)
	}
}

func TestAllIter_ResetOnSecondCall(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var b comp.Blueprint
	b.Init(cc, comp.Track(new(iter.Col[iterPos])))
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()

	v := NewView(cat, comp.Track(new(iter.Col[iterPos])))

	count := func() int {
		n := 0
		v.All()
		for v.Next() {
			n += len(v.Cursor.IDs)
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
