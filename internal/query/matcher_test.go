package query

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
	"github.com/kjkrol/goke/iter"
)

func TestMatcher_Clear(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.Col[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	m := NewMatcher(cat, _trackOpt0)
	if len(m.BakedTables) == 0 {
		t.Fatal("matcher should have baked tables before Clear")
	}

	m.Clear()

	if m.BakedTables != nil {
		t.Error("BakedTables should be nil after Clear")
	}
	if m.EntityIndex != nil {
		t.Error("EntityIndex should be nil after Clear")
	}
}

func TestMatcher_AllAfterClear(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.Col[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	m := NewMatcher(cat, _trackOpt0)
	m.Clear()

	it := m.All()
	if it.Next() {
		t.Error("All() on cleared matcher should yield no results")
	}
}

func TestMatcher_InitPanicsOnConflict(t *testing.T) {
	cat, _, _ := newQueryCatalog()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when same component is both required and excluded")
		}
	}()

	NewMatcher(cat, comp.Include[iterPos](), comp.Exclude[iterPos]())
}

func TestCatalog_NewMatcher(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.Col[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	_trackOpt0 := comp.Track(new(iter.Col[iterPos]))
	m := NewMatcher(cat, _trackOpt0)

	if len(m.BakedTables) != 1 {
		t.Errorf("expected 1 BakedTable, got %d", len(m.BakedTables))
	}
}

func TestCatalog_Reset(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.Col[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	m := NewMatcher(cat, comp.Track(new(iter.Col[iterPos])))
	if len(m.BakedTables) == 0 {
		t.Fatal("matcher should have baked tables before Reset")
	}

	cat.Reset()

	if len(m.BakedTables) != 0 {
		t.Errorf("expected 0 BakedTables after catalog Reset, got %d", len(m.BakedTables))
	}
}

func TestCatalog_AddPanicsWhenFull(t *testing.T) {
	var cc comp.DefIndex
	cc.Init()
	var em ent.Manager
	cat := new(Catalog)
	cat.Init(&cc, &em.AddressBook.Index, &em.ArchCatalog, Config{Cap: 2})
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
