package ent_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

func TestFactory_SingleBatchSingleChunk(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	var posCol iter.ArrayRef[Position]
	var spec comp.AccessSpec
	spec.Init(&mi, comp.Track(&posCol))

	factory := m.CreateFactory(spec)
	factory.Create(5)

	if !factory.Next() {
		t.Fatal("expected Next to report a batch")
	}
	if len(factory.IDs) != 5 {
		t.Fatalf("expected 5 IDs, got %d", len(factory.IDs))
	}

	seen := make(map[uint64]bool, 5)
	for _, id := range factory.IDs {
		if seen[uint64(id)] {
			t.Errorf("duplicate ID %v", id)
		}
		seen[uint64(id)] = true
	}

	positions := posCol.Slice(&factory.Cursor)
	for i := range positions {
		positions[i] = Position{X: float64(i), Y: float64(i)}
	}
	for i, p := range posCol.Slice(&factory.Cursor) {
		if p.X != float64(i) {
			t.Errorf("entity %d: expected X=%d, got %v", i, i, p.X)
		}
	}

	if factory.Next() {
		t.Error("expected a second Next call to report no more batches")
	}
}

func TestFactory_MultipleBatchesAcrossChunks(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	var posCol iter.ArrayRef[Position]
	var spec comp.AccessSpec
	spec.Init(&mi, comp.Track(&posCol))

	factory := m.CreateFactory(spec)

	// Large enough to force at least one chunk boundary regardless of
	// component size / L1 cache geometry.
	const count = 5000
	factory.Create(count)

	total := 0
	seen := make(map[uint64]bool, count)
	for factory.Next() {
		positions := posCol.Slice(&factory.Cursor)
		for i, id := range factory.IDs {
			if seen[uint64(id)] {
				t.Fatalf("duplicate ID %v", id)
			}
			seen[uint64(id)] = true
			positions[i] = Position{X: float64(total + i)}
		}
		total += len(factory.IDs)
	}

	if total != count {
		t.Errorf("expected %d entities created, got %d", count, total)
	}
}

func TestFactory_ZeroCount(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	var posCol iter.ArrayRef[Position]
	var spec comp.AccessSpec
	spec.Init(&mi, comp.Track(&posCol))

	factory := m.CreateFactory(spec)
	factory.Create(0)

	if factory.Next() {
		t.Error("expected Next to immediately report no batches for Create(0)")
	}
	if factory.IDs != nil {
		t.Errorf("expected IDs to be nil after an exhausted Next, got %v", factory.IDs)
	}
}

func TestFactory_ReuseAcrossCreateCalls(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	var posCol iter.ArrayRef[Position]
	var spec comp.AccessSpec
	spec.Init(&mi, comp.Track(&posCol))

	factory := m.CreateFactory(spec)

	factory.Create(2)
	factory.Next()
	firstBatch := append([]uid.UID64{}, factory.IDs...)

	factory.Create(2)
	factory.Next()
	secondBatch := factory.IDs

	for _, a := range firstBatch {
		for _, b := range secondBatch {
			if a == b {
				t.Errorf("expected distinct IDs across Create calls, got %v in both batches", a)
			}
		}
	}
}

func TestFactory_TaggedOnlyArchetype(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(3)

	total := 0
	for factory.Next() {
		total += len(factory.IDs)
	}
	if total != 3 {
		t.Errorf("expected 3 entities, got %d", total)
	}
}
