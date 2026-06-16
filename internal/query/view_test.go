package query

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
)

func TestView_Clear(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	_, _ = em.UpsertComp(e, posMeta)

	v := NewView(cat, comp.Track[iterPos]())
	if len(v.BakedTables) == 0 {
		t.Fatal("view should have baked tables before Clear")
	}

	v.Clear()

	if v.BakedTables != nil {
		t.Error("BakedTables should be nil after Clear")
	}
	if v.EntityIndex != nil {
		t.Error("EntityIndex should be nil after Clear")
	}
}

func TestView_AllAfterClear(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	_, _ = em.UpsertComp(e, posMeta)

	v := NewView(cat, comp.Track[iterPos]())
	v.Clear()

	it := v.All()
	if it.Next() {
		t.Error("All() on cleared view should yield no results")
	}
}

func TestView_InitPanicsOnConflict(t *testing.T) {
	cat, _, _ := newQueryCatalog()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when same component is both required and excluded")
		}
	}()

	NewView(cat, comp.Include[iterPos](), comp.Exclude[iterPos]())
}

func TestCatalog_NewView(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	_, _ = em.UpsertComp(e, posMeta)

	v := NewView(cat, comp.Track[iterPos]())

	if len(v.BakedTables) != 1 {
		t.Errorf("expected 1 BakedTable, got %d", len(v.BakedTables))
	}
}

func TestCatalog_Reset(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	posMeta := cc.Intern(reflect.TypeFor[iterPos]())

	e := em.Create()
	_, _ = em.UpsertComp(e, posMeta)

	v := NewView(cat, comp.Track[iterPos]())
	if len(v.BakedTables) == 0 {
		t.Fatal("view should have baked tables before Reset")
	}

	cat.Reset()

	if len(v.BakedTables) != 0 {
		t.Errorf("expected 0 BakedTables after catalog Reset, got %d", len(v.BakedTables))
	}
}

func TestCatalog_AddPanicsWhenFull(t *testing.T) {
	var cc comp.Catalog
	cc.Init()
	var em ent.Manager
	cat := new(Catalog)
	cat.Init(&cc, &em, Config{Cap: 2})
	em.Init(ent.DefaultConfig(), cat.OnArchetypeCreated)

	cat.Add()
	cat.Add()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when catalog capacity is exceeded")
		}
	}()

	cat.Add()
}
