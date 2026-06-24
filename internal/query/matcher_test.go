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
	accessSpec.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
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
	accessSpec.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
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

// A non-conflicting Exclude must be applied normally (no panic) and take
// effect — distinct from TestMatcher_InitPanicsOnConflict, which never
// reaches the exclude-mask-set line at all since it panics first.
func TestMatcher_Init_AppliesNonConflictingExclude(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	spec := comp.AccessSpec{}
	spec.Init(cc, comp.Track(&pos))
	f := em.CreateFactory(spec)
	f.Create(1)
	f.Next()

	m := NewMatcher(cat, comp.Track(new(iter.ArrayRef[iterPos])), comp.Exclude[iterVel]())

	count := 0
	m.All()
	for m.Next() {
		count += len(m.Cursor.IDs)
	}
	if count != 1 {
		t.Errorf("expected the Pos-only entity to match (no Vel to exclude), got %d", count)
	}
}

// NewMatcher must panic when applying its own opts fails — distinct from
// Matcher.Init's panic, which only fires for an Include/Exclude conflict
// inside Init itself.
func TestNewMatcher_PanicsOnOptError(t *testing.T) {
	cat, _, _ := newQueryCatalog()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected NewMatcher to panic when an opt returns an error")
		}
	}()

	NewMatcher(cat, comp.Include[iterPos](), comp.Include[iterPos]())
}

// Next on a Matcher that never had All or Pick called must return false
// rather than iterate anything.
func TestMatcher_Next_WithoutModeReturnsFalse(t *testing.T) {
	cat, _, _ := newQueryCatalog()
	m := NewMatcher(cat)

	if m.Next() {
		t.Error("expected Next to return false when no mode was set")
	}
}

func TestCatalog_NewMatcher(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	_trackOpt0 := comp.Track(new(iter.ArrayRef[iterPos]))
	m := NewMatcher(cat, _trackOpt0)

	if len(m.BakedTables) != 1 {
		t.Errorf("expected 1 BakedTable, got %d", len(m.BakedTables))
	}
}

func TestCatalog_Reset(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()

	m := NewMatcher(cat, comp.Track(new(iter.ArrayRef[iterPos])))
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
