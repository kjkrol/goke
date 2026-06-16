package query

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
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
	if v.EntSlice != nil {
		t.Error("Chunk should be nil after exhaustion")
	}
}

func TestAllIter_SingleEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	ptr, _ := em.UpsertComp(e, posMeta)
	*(*iterPos)(ptr) = iterPos{X: 1, Y: 2}

	var pos Col[iterPos]
	v := NewView(cat, pos.Track())

	var visited []uid.UID64
	v.All()
	for v.Next() {
		posSlice := pos.Slice(v)
		for i, entity := range v.EntSlice {
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
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e1 := em.Create()
	ptr1, _ := em.UpsertComp(e1, posMeta)
	*(*iterPos)(ptr1) = iterPos{X: 10, Y: 20}

	e2 := em.Create()
	ptr2, _ := em.UpsertComp(e2, posMeta)
	*(*iterPos)(ptr2) = iterPos{X: 30, Y: 40}

	var pos Col[iterPos]
	v := NewView(cat, pos.Track())

	got := map[uid.UID64]iterPos{}
	v.All()
	for v.Next() {
		posSlice := pos.Slice(v)
		for i, entity := range v.EntSlice {
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
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	ptr, _ := em.UpsertComp(e, posMeta)
	*(*iterPos)(ptr) = iterPos{X: 5, Y: 3}

	var pos Col[iterPos]
	v := NewView(cat, pos.Track())

	v.All()
	for v.Next() {
		posSlice := pos.Slice(v)
		for i := range v.EntSlice {
			posSlice[i].X += posSlice[i].Y
		}
	}

	// re-read from live memory via a second pass
	v.All()
	for v.Next() {
		posSlice := pos.Slice(v)
		for range v.EntSlice {
			if posSlice[0].X != 8 {
				t.Errorf("expected X=8 after mutation, got %v", posSlice[0].X)
			}
		}
	}
}

func TestAllIter_MultipleArchetypes(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())
	velMeta := cc.Intern(reflect.TypeFor[iterVel]())

	// archetype A: pos only
	eA := em.Create()
	pA, _ := em.UpsertComp(eA, posMeta)
	*(*iterPos)(pA) = iterPos{X: 1}

	// archetype B: pos + vel
	eB := em.Create()
	pB, _ := em.UpsertComp(eB, posMeta)
	*(*iterPos)(pB) = iterPos{X: 2}
	_, _ = em.UpsertComp(eB, velMeta)

	v := NewView(cat, comp.Track[iterPos]())

	visited := map[uid.UID64]bool{}
	v.All()
	for v.Next() {
		for _, e := range v.EntSlice {
			visited[e] = true
		}
	}

	if !visited[eA] || !visited[eB] {
		t.Errorf("expected both entities visited, got %v", visited)
	}
}

func TestAllIter_ResetOnSecondCall(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	_, _ = em.UpsertComp(e, posMeta)

	v := NewView(cat, comp.Track[iterPos]())

	count := func() int {
		n := 0
		v.All()
		for v.Next() {
			n += len(v.EntSlice)
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
