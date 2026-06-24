package query

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/iter"
	"github.com/kjkrol/uid"
)

func TestFilterIter_EmptySelected(t *testing.T) {
	cat, _, _ := newQueryCatalog()
	m := NewMatcher(cat)

	m.Pick(nil)
	if m.Next() {
		t.Error("Next() with nil selected should return false")
	}
}

func TestFilterIter_EntityInMatcher(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, _trackOpt0)
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	m := NewMatcher(cat, _trackOpt0)

	m.Pick([]uid.UID64{e})
	if !m.Next() {
		t.Fatal("expected Next() true for valid entity")
	}
	if m.Entity != e {
		t.Errorf("expected Entity %v, got %v", e, m.Entity)
	}
	if m.Idx != 0 {
		t.Errorf("expected Idx 0, got %d", m.Idx)
	}
	if m.Next() {
		t.Error("expected no more entities")
	}
}

func TestFilterIter_EntityNotInMatcher(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	// entity has vel only — matcher requires pos
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.ArrayRef[iterVel])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
	m := NewMatcher(cat, _trackOpt0)

	m.Pick([]uid.UID64{e})
	if m.Next() {
		t.Error("entity not matching matcher mask should be skipped")
	}
}

func TestFilterIter_SkipsDeletedEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, _trackOpt0)
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	m := NewMatcher(cat, _trackOpt0)

	em.Remove(e)

	m.Pick([]uid.UID64{e})
	if m.Next() {
		t.Error("deleted entity should be skipped")
	}
}

func TestFilterIter_At(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	posOpt := comp.Track(&pos)
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, posOpt)
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 5, Y: 7}

	m := NewMatcher(cat, posOpt)

	m.Pick([]uid.UID64{e})
	if !m.Next() {
		t.Fatal("expected Next() true")
	}

	p := pos.At(&m.Cursor)
	if p == nil {
		t.Fatal("At returned nil")
	}
	if *p != (iterPos{X: 5, Y: 7}) {
		t.Errorf("expected {5,7}, got %v", *p)
	}
}

func TestFilterIter_AtMutation(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	posOpt := comp.Track(&pos)
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, posOpt)
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 1}

	m := NewMatcher(cat, posOpt)

	// mutate via At pointer
	m.Pick([]uid.UID64{e})
	if !m.Next() {
		t.Fatal("expected Next() true")
	}
	pos.At(&m.Cursor).X = 99

	// read back in a second pass
	m.Pick([]uid.UID64{e})
	if !m.Next() {
		t.Fatal("expected Next() true on second pass")
	}
	if pos.At(&m.Cursor).X != 99 {
		t.Error("mutation via At pointer was not persisted")
	}
}

func TestFilterIter_IdxTracksPosition(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, _trackOpt0)
	f := em.CreateFactory(accessSpec)

	f.Create(1)
	f.Next()
	e0 := f.IDs[0]

	f.Create(1)
	f.Next()
	e1 := f.IDs[0]

	m := NewMatcher(cat, _trackOpt0)

	// position 0 = e0 (valid), position 1 = ghost (invalid), position 2 = e1 (valid)
	ghost := uid.UID64(0xDEADBEEF)
	selected := []uid.UID64{e0, ghost, e1}

	var idxs []int
	var entities []uid.UID64
	m.Pick(selected)
	for m.Next() {
		idxs = append(idxs, m.Idx)
		entities = append(entities, m.Entity)
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

	var pos iter.ArrayRef[iterPos]
	var vel iter.ArrayRef[iterVel]
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(&pos), comp.Track(&vel))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 3, Y: 4}
	*vel.At(&f.Cursor) = iterVel{VX: 1, VY: 2}

	m := NewMatcher(cat, comp.Track(&pos), comp.Track(&vel))

	m.Pick([]uid.UID64{e})
	if !m.Next() {
		t.Fatal("expected Next() true")
	}

	if *pos.At(&m.Cursor) != (iterPos{X: 3, Y: 4}) {
		t.Errorf("pos: want {3,4} got %v", *pos.At(&m.Cursor))
	}
	if *vel.At(&m.Cursor) != (iterVel{VX: 1, VY: 2}) {
		t.Errorf("vel: want {1,2} got %v", *vel.At(&m.Cursor))
	}
}

func TestFilterIter_ResetOnSecondCall(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]

	m := NewMatcher(cat, comp.Track(new(iter.ArrayRef[iterPos])))
	selected := []uid.UID64{e}

	count := func() int {
		n := 0
		m.Pick(selected)
		for m.Next() {
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
