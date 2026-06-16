package query

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
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
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	ptr, _ := em.UpsertComp(e, posMeta)
	*(*iterPos)(ptr) = iterPos{X: 5, Y: 7}

	v := NewView(cat, comp.Track[iterPos]())

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
	_ = cc.Intern(reflect.TypeFor[iterPos]())
	velMeta := cc.Intern(reflect.TypeFor[iterVel]())

	// entity has vel only — view requires pos
	e := em.Create()
	_, _ = em.UpsertComp(e, velMeta)

	v := NewView(cat, comp.Track[iterPos]())

	v.Filter([]uid.UID64{e})
	if v.Next() {
		t.Error("entity not matching view mask should be skipped")
	}
}

func TestFilterIter_SkipsDeletedEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	_, _ = em.UpsertComp(e, posMeta)

	v := NewView(cat, comp.Track[iterPos]())

	em.Remove(e)

	v.Filter([]uid.UID64{e})
	if v.Next() {
		t.Error("deleted entity should be skipped")
	}
}

func TestFilterIter_At(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	ptr, _ := em.UpsertComp(e, posMeta)
	*(*iterPos)(ptr) = iterPos{X: 5, Y: 7}

	var pos Col[iterPos]
	v := NewView(cat, pos.Track())

	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true")
	}

	p := pos.At(v)
	if p == nil {
		t.Fatal("At returned nil")
	}
	if *p != (iterPos{X: 5, Y: 7}) {
		t.Errorf("expected {5,7}, got %v", *p)
	}
}

func TestFilterIter_AtMutation(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	ptr, _ := em.UpsertComp(e, posMeta)
	*(*iterPos)(ptr) = iterPos{X: 1}

	var pos Col[iterPos]
	v := NewView(cat, pos.Track())

	// mutate via At pointer
	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true")
	}
	pos.At(v).X = 99

	// read back in a second pass
	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true on second pass")
	}
	if pos.At(v).X != 99 {
		t.Error("mutation via At pointer was not persisted")
	}
}

func TestFilterIter_IdxTracksPosition(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e0 := em.Create()
	_, _ = em.UpsertComp(e0, posMeta)
	e1 := em.Create()
	_, _ = em.UpsertComp(e1, posMeta)

	v := NewView(cat, comp.Track[iterPos]())

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
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())
	velMeta := cc.Intern(reflect.TypeFor[iterVel]())

	e := em.Create()
	pp, _ := em.UpsertComp(e, posMeta)
	*(*iterPos)(pp) = iterPos{X: 3, Y: 4}
	vp, _ := em.UpsertComp(e, velMeta)
	*(*iterVel)(vp) = iterVel{VX: 1, VY: 2}

	var pos Col[iterPos]
	var vel Col[iterVel]
	v := NewView(cat, pos.Track(), vel.Track())

	v.Filter([]uid.UID64{e})
	if !v.Next() {
		t.Fatal("expected Next() true")
	}

	if *pos.At(v) != (iterPos{X: 3, Y: 4}) {
		t.Errorf("pos: want {3,4} got %v", *pos.At(v))
	}
	if *vel.At(v) != (iterVel{VX: 1, VY: 2}) {
		t.Errorf("vel: want {1,2} got %v", *vel.At(v))
	}
}

func TestFilterIter_ResetOnSecondCall(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	_, _ = em.UpsertComp(e, posMeta)

	v := NewView(cat, comp.Track[iterPos]())
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
