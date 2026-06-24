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

// TestBakedTablesCatalog_GrowDoesNotOverAllocate guards against a regression
// where grow() re-allocated and doubled an already-inflated cap() on every
// single call instead of reusing existing spare capacity. Because Add() is
// called once per sequentially-numbered archetype ID, that bug compounded
// into exponential growth (cap doubling on every +1 archetype), reaching
// hundreds of millions of elements — and eventually an out-of-memory crash —
// after only a few dozen archetypes.
func TestBakedTablesCatalog_GrowDoesNotOverAllocate(t *testing.T) {
	compCatalog := newDefIndex()
	archCatalog := setupArchCatalog()

	const archetypeCount = 40

	var c BakedTablesCatalog
	for i := range archetypeCount {
		// reflect.StructOf gives each iteration a distinct reflect.Type, so
		// each one interns as a separate component and therefore lands in
		// its own archetype with a sequentially-assigned archetype ID.
		fieldType := reflect.StructOf([]reflect.StructField{
			{Name: "Marker", Type: reflect.ArrayOf(i+1, reflect.TypeFor[byte]())},
		})
		meta := compCatalog.Intern(fieldType)
		archID := archCatalog.Upsert(comp.Composition{}.With(meta))
		c.Add(&archCatalog.Archetypes[archID], []comp.ID{meta.ID})
	}

	if len(c.BakedTables) != archetypeCount {
		t.Fatalf("expected %d BakedTables, got %d", archetypeCount, len(c.BakedTables))
	}

	// A correctly amortized grow() never needs more than a small constant
	// multiple of the actual element count. The buggy version reached
	// cap() in the hundreds of millions for this same input.
	if gotCap := cap(c.archTableIndex); gotCap > archetypeCount*4 {
		t.Errorf("archTableIndex cap = %d, want <= %d (exponential over-allocation regression)", gotCap, archetypeCount*4)
	}
}
