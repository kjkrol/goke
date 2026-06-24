package query

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/comp"
)

type (
	testPos struct{ X, Y float32 }
	testVel struct{ VX, VY float32 }
)

func newDefIndex() comp.DefIndex {
	var c comp.DefIndex
	c.Init()
	return c
}

func setupArchCatalog() *arch.Catalog {
	catalog := &arch.Catalog{}
	catalog.Init(func(*arch.Archetype) {})
	return catalog
}

func TestBakedTablesCatalog_EmptyInitially(t *testing.T) {
	var c BakedTablesCatalog

	if c.BakedTables != nil {
		t.Error("expected nil BakedTables initially")
	}
	if got := c.Get(arch.ID(1)); got != nil {
		t.Error("expected nil from Get on empty catalog")
	}
	if got := c.Get(arch.ID(0)); got != nil {
		t.Error("expected nil from Get with zero archID on empty catalog")
	}
}

func TestBakedTablesCatalog_AddAndGet(t *testing.T) {
	compCatalog := newDefIndex()
	posMeta := compCatalog.Intern(reflect.TypeFor[testPos]())

	archCatalog := setupArchCatalog()
	archID := archCatalog.Upsert(comp.Composition{}.With(posMeta))
	archetype := &archCatalog.Archetypes[archID]

	var c BakedTablesCatalog
	c.Add(archetype, []comp.ID{posMeta.ID})

	bt := c.Get(archID)
	if bt == nil {
		t.Fatal("expected BakedTable, got nil")
	}
	if bt.Table != &archetype.Table {
		t.Error("BakedTable points to wrong table")
	}
	if len(bt.CompOffsets) != 1 {
		t.Errorf("expected 1 CompOffset, got %d", len(bt.CompOffsets))
	}
}

func TestBakedTablesCatalog_GetOutOfRange(t *testing.T) {
	compCatalog := newDefIndex()
	posMeta := compCatalog.Intern(reflect.TypeFor[testPos]())

	archCatalog := setupArchCatalog()
	archID := archCatalog.Upsert(comp.Composition{}.With(posMeta))
	archetype := &archCatalog.Archetypes[archID]

	var c BakedTablesCatalog
	c.Add(archetype, []comp.ID{posMeta.ID})

	if got := c.Get(arch.ID(999)); got != nil {
		t.Error("expected nil for archID beyond mapping range")
	}
}

func TestBakedTablesCatalog_GetNonMatchedArchID(t *testing.T) {
	compCatalog := newDefIndex()
	posMeta := compCatalog.Intern(reflect.TypeFor[testPos]())
	velMeta := compCatalog.Intern(reflect.TypeFor[testVel]())

	archCatalog := setupArchCatalog()
	archID1 := archCatalog.Upsert(comp.Composition{}.With(posMeta))
	archID2 := archCatalog.Upsert(comp.Composition{}.With(velMeta))
	archetype1 := &archCatalog.Archetypes[archID1]

	var c BakedTablesCatalog
	c.Add(archetype1, []comp.ID{posMeta.ID})

	// archID2 was never added to the catalog
	if got := c.Get(archID2); got != nil {
		t.Errorf("expected nil for archID %d not added to catalog", archID2)
	}
}

func TestBakedTablesCatalog_MultipleArchetypes(t *testing.T) {
	compCatalog := newDefIndex()
	posMeta := compCatalog.Intern(reflect.TypeFor[testPos]())
	velMeta := compCatalog.Intern(reflect.TypeFor[testVel]())

	archCatalog := setupArchCatalog()
	archID1 := archCatalog.Upsert(comp.Composition{}.With(posMeta))
	archID2 := archCatalog.Upsert(comp.Composition{}.With(velMeta))
	arch1 := &archCatalog.Archetypes[archID1]
	arch2 := &archCatalog.Archetypes[archID2]

	var c BakedTablesCatalog
	c.Add(arch1, []comp.ID{posMeta.ID})
	c.Add(arch2, []comp.ID{velMeta.ID})

	bt1 := c.Get(archID1)
	bt2 := c.Get(archID2)

	if bt1 == nil || bt2 == nil {
		t.Fatal("expected both BakedTables to be non-nil")
	}
	if bt1 == bt2 {
		t.Error("expected distinct BakedTable pointers")
	}
	if bt1.Table != &arch1.Table {
		t.Error("bt1 points to wrong table")
	}
	if bt2.Table != &arch2.Table {
		t.Error("bt2 points to wrong table")
	}
}

func TestBakedTablesCatalog_Clear(t *testing.T) {
	compCatalog := newDefIndex()
	posMeta := compCatalog.Intern(reflect.TypeFor[testPos]())

	archCatalog := setupArchCatalog()
	archID := archCatalog.Upsert(comp.Composition{}.With(posMeta))
	archetype := &archCatalog.Archetypes[archID]

	var c BakedTablesCatalog
	c.Add(archetype, []comp.ID{posMeta.ID})

	c.Clear()

	if c.BakedTables != nil {
		t.Error("expected nil BakedTables after Clear")
	}
	if got := c.Get(archID); got != nil {
		t.Error("expected nil from Get after Clear")
	}
}

func TestBakedTablesCatalog_GrowOnSequentialArchIDs(t *testing.T) {
	compCatalog := newDefIndex()
	posMeta := compCatalog.Intern(reflect.TypeFor[testPos]())
	velMeta := compCatalog.Intern(reflect.TypeFor[testVel]())
	type Tag struct{}
	tagMeta := compCatalog.Intern(reflect.TypeFor[Tag]())

	archCatalog := setupArchCatalog()
	archID1 := archCatalog.Upsert(comp.Composition{}.With(posMeta))
	archID2 := archCatalog.Upsert(comp.Composition{}.With(velMeta))
	archID3 := archCatalog.Upsert(comp.Composition{}.With(tagMeta))

	var c BakedTablesCatalog
	c.Add(&archCatalog.Archetypes[archID1], []comp.ID{posMeta.ID})
	c.Add(&archCatalog.Archetypes[archID2], []comp.ID{velMeta.ID})
	c.Add(&archCatalog.Archetypes[archID3], []comp.ID{tagMeta.ID})

	if c.Get(archID1) == nil {
		t.Error("expected BakedTable for archID1")
	}
	if c.Get(archID2) == nil {
		t.Error("expected BakedTable for archID2")
	}
	if c.Get(archID3) == nil {
		t.Error("expected BakedTable for archID3")
	}
	if len(c.BakedTables) != 3 {
		t.Errorf("expected 3 BakedTables, got %d", len(c.BakedTables))
	}
}
