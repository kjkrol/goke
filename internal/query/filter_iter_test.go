package query

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
	"github.com/kjkrol/uid"
)

func TestFilterIter_EmptySelected(t *testing.T) {
	cat, _, _ := newQueryCatalog()
	v := NewView(cat)

	v.Filter(nil)
	if v.Next() {
		t.Error("Next() with nil selected should return false")
	}
}

func TestFilterIter_EntityInView(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	var b comp.Blueprint
	b.Init(cc, _trackOpt0)
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	v := NewView(cat, _trackOpt0)

	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true for valid entity")
	}
	if v.Entity != e {
		t.Errorf("expected Entity %v, got %v", e, v.Entity)
	}
	if v.Idx != 0 {
		t.Errorf("expected Idx 0, got %d", v.Idx)
	}
	if v.Next() {
		t.Error("expected no more entities")
	}
}

func TestFilterIter_EntityNotInView(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	// entity has vel only — view requires pos
	var b comp.Blueprint
	b.Init(cc, comp.Track(new(iter.Col[iterVel])))
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	v := NewView(cat, _trackOpt0)

	v.Filter([]uid.UID64{e})
	if v.Next() {
		t.Error("entity not matching view mask should be skipped")
	}
}

func TestFilterIter_SkipsDeletedEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	var b comp.Blueprint
	b.Init(cc, _trackOpt0)
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	v := NewView(cat, _trackOpt0)

	em.Remove(e)

	v.Filter([]uid.UID64{e})
	if v.Next() {
		t.Error("deleted entity should be skipped")
	}
}

func TestFilterIter_At(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.Col[iterPos]
	posOpt := comp.Track(&pos)
	var b comp.Blueprint
	b.Init(cc, posOpt)
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 5, Y: 7}

	v := NewView(cat, posOpt)

	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true")
	}

	p := pos.At(&v.Cursor)
	if p == nil {
		t.Fatal("At returned nil")
	}
	if *p != (iterPos{X: 5, Y: 7}) {
		t.Errorf("expected {5,7}, got %v", *p)
	}
}

func TestFilterIter_AtMutation(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.Col[iterPos]
	posOpt := comp.Track(&pos)
	var b comp.Blueprint
	b.Init(cc, posOpt)
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 1}

	v := NewView(cat, posOpt)

	// mutate via At pointer
	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true")
	}
	pos.At(&v.Cursor).X = 99

	// read back in a second pass
	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true on second pass")
	}
	if pos.At(&v.Cursor).X != 99 {
		t.Error("mutation via At pointer was not persisted")
	}
}

func TestFilterIter_IdxTracksPosition(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	var b comp.Blueprint
	b.Init(cc, _trackOpt0)
	f := em.CreateFactory(b)

	f.Create(1)
	f.Next()
	e0 := f.IDs[0]

	f.Create(1)
	f.Next()
	e1 := f.IDs[0]

	v := NewView(cat, _trackOpt0)

	// position 0 = e0 (valid), position 1 = ghost (invalid), position 2 = e1 (valid)
	ghost := uid.UID64(0xDEADBEEF)
	selected := []uid.UID64{e0, ghost, e1}

	var idxs []int
	var entities []uid.UID64
	v.Filter(selected)
	for v.Next() {
		idxs = append(idxs, v.Idx)
		entities = append(entities, v.Entity)
	}

	if len(idxs) != 2 {
		t.Fatalf("expected 2 results, got %d", len(idxs))
	}
	if idxs[0] != 0 || idxs[1] != 2 {
		t.Errorf("expected Idx [0,2], got %v", idxs)
	}
	if entities[0] != e0 || entities[1] != e1 {
		t.Errorf("unexpected entities: %v", entities)
	}
}

func TestFilterIter_MultipleComps(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.Col[iterPos]
	var vel iter.Col[iterVel]
	var b comp.Blueprint
	b.Init(cc, comp.Track(&pos), comp.Track(&vel))
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 3, Y: 4}
	*vel.At(&f.Cursor) = iterVel{VX: 1, VY: 2}

	v := NewView(cat, comp.Track(&pos), comp.Track(&vel))

	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true")
	}

	if *pos.At(&v.Cursor) != (iterPos{X: 3, Y: 4}) {
		t.Errorf("pos: want {3,4} got %v", *pos.At(&v.Cursor))
	}
	if *vel.At(&v.Cursor) != (iterVel{VX: 1, VY: 2}) {
		t.Errorf("vel: want {1,2} got %v", *vel.At(&v.Cursor))
	}
}

func TestFilterIter_ResetOnSecondCall(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var b comp.Blueprint
	b.Init(cc, comp.Track(new(iter.Col[iterPos])))
	f := em.CreateFactory(b)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	v := NewView(cat, comp.Track(new(iter.Col[iterPos])))
	selected := []uid.UID64{e}

	count := func() int {
		n := 0
		v.Filter(selected)
		for v.Next() {
			n++
		}
		return n
	}

	if n := count(); n != 1 {
		t.Errorf("first Filter() call: expected 1 entity, got %d", n)
	}
	if n := count(); n != 1 {
		t.Errorf("second Filter() call: expected 1 entity, got %d", n)
	}
}
