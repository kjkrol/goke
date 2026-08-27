package query

import (
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
	"github.com/kjkrol/uid"
)

func TestOptional_AllMode_PresentAndAbsent(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	// archetype A: pos only — Vel is absent.
	var posA iter.ArrayRef[iterPos]
	specA := comp.AccessSpec{}
	specA.Init(cc, comp.Track(&posA))
	fA := em.CreateFactory(specA)
	fA.Create(1)
	fA.Next()
	eA := fA.IDs[0]

	// archetype B: pos + vel — Vel is present.
	var posB iter.ArrayRef[iterPos]
	var velB iter.ArrayRef[iterVel]
	specB := comp.AccessSpec{}
	specB.Init(cc, comp.Track(&posB), comp.Track(&velB))
	fB := em.CreateFactory(specB)
	fB.Create(1)
	fB.Next()
	eB := fB.IDs[0]
	*velB.At(&fB.Cursor) = iterVel{VX: 7, VY: 8}

	var pos iter.ArrayRef[iterPos]
	var vel iter.OptArrayRef[iterVel]
	m := NewMatcher(cat, comp.Track(&pos), comp.Optional(&vel))

	seen := map[uid.UID64]bool{}
	m.All()
	for m.Next() {
		present := vel.Present(&m.Cursor)
		for i, e := range m.Cursor.IDs {
			seen[e] = true
			switch e {
			case eA:
				if present {
					t.Errorf("expected eA's chunk to report Vel absent")
				}
				if got := vel.Slice(&m.Cursor); got != nil {
					t.Errorf("expected nil Slice for eA's chunk, got %v", got)
				}
				if got := vel.At(&m.Cursor); got != nil {
					t.Errorf("expected nil At for eA, got %v", got)
				}
			case eB:
				if !present {
					t.Errorf("expected eB's chunk to report Vel present")
				}
				if got := vel.Slice(&m.Cursor)[i]; got != (iterVel{VX: 7, VY: 8}) {
					t.Errorf("expected {7,8}, got %+v", got)
				}
			}
		}
	}

	if !seen[eA] || !seen[eB] {
		t.Fatalf("expected both entities matched (Optional never excludes), got %v", seen)
	}
}

func TestOptional_PickMode_PresentAndAbsent(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var posA iter.ArrayRef[iterPos]
	specA := comp.AccessSpec{}
	specA.Init(cc, comp.Track(&posA))
	fA := em.CreateFactory(specA)
	fA.Create(1)
	fA.Next()
	eA := fA.IDs[0]

	var posB iter.ArrayRef[iterPos]
	var velB iter.ArrayRef[iterVel]
	specB := comp.AccessSpec{}
	specB.Init(cc, comp.Track(&posB), comp.Track(&velB))
	fB := em.CreateFactory(specB)
	fB.Create(1)
	fB.Next()
	eB := fB.IDs[0]
	*velB.At(&fB.Cursor) = iterVel{VX: 3, VY: 4}

	var pos iter.ArrayRef[iterPos]
	var vel iter.OptArrayRef[iterVel]
	m := NewMatcher(cat, comp.Track(&pos), comp.Optional(&vel))

	m.Pick([]uid.UID64{eA, eB})
	for m.Next() {
		switch m.Entity {
		case eA:
			if vel.Present(&m.Cursor) {
				t.Error("expected eA to report Vel absent in Pick mode")
			}
			if got := vel.At(&m.Cursor); got != nil {
				t.Errorf("expected nil At for eA, got %v", got)
			}
		case eB:
			if !vel.Present(&m.Cursor) {
				t.Error("expected eB to report Vel present in Pick mode")
			}
			if got := vel.At(&m.Cursor); got == nil || *got != (iterVel{VX: 3, VY: 4}) {
				t.Errorf("expected {3,4}, got %+v", got)
			}
		default:
			t.Fatalf("unexpected entity %v", m.Entity)
		}
	}
}

func TestOptional_Seek_PresentAndAbsent(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var posA iter.ArrayRef[iterPos]
	specA := comp.AccessSpec{}
	specA.Init(cc, comp.Track(&posA))
	fA := em.CreateFactory(specA)
	fA.Create(1)
	fA.Next()
	eA := fA.IDs[0]

	var posB iter.ArrayRef[iterPos]
	var velB iter.ArrayRef[iterVel]
	specB := comp.AccessSpec{}
	specB.Init(cc, comp.Track(&posB), comp.Track(&velB))
	fB := em.CreateFactory(specB)
	fB.Create(1)
	fB.Next()
	eB := fB.IDs[0]
	*velB.At(&fB.Cursor) = iterVel{VX: 5, VY: 6}

	var pos iter.ArrayRef[iterPos]
	var vel iter.OptArrayRef[iterVel]
	m := NewMatcher(cat, comp.Track(&pos), comp.Optional(&vel))

	if !m.Seek(eA) {
		t.Fatal("expected Seek to find eA")
	}
	if vel.Present(&m.Cursor) {
		t.Error("expected eA to report Vel absent via Seek")
	}

	if !m.Seek(eB) {
		t.Fatal("expected Seek to find eB")
	}
	if !vel.Present(&m.Cursor) {
		t.Fatal("expected eB to report Vel present via Seek")
	}
	if got := vel.At(&m.Cursor); got == nil || *got != (iterVel{VX: 5, VY: 6}) {
		t.Errorf("expected {5,6}, got %+v", got)
	}

	// Re-seeking eA after eB must not leak eB's optional offsets/presence.
	if !m.Seek(eA) {
		t.Fatal("expected re-Seek to find eA")
	}
	if vel.Present(&m.Cursor) {
		t.Error("expected eA to report Vel absent again after seeking eB in between")
	}
}

func TestOptional_SeekH_SameArchetypeReusesCache(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	var velCol iter.ArrayRef[iterVel]
	spec := comp.AccessSpec{}
	spec.Init(cc, comp.Track(&pos), comp.Track(&velCol))
	f := em.CreateFactory(spec)
	f.Create(2)
	f.Next()
	e0, e1 := f.IDs[0], f.IDs[1]
	*velCol.At(&f.Cursor) = iterVel{VX: 1, VY: 1}
	f.Cursor.Slot = 1
	*velCol.At(&f.Cursor) = iterVel{VX: 2, VY: 2}

	var trackedPos iter.ArrayRef[iterPos]
	var vel iter.OptArrayRef[iterVel]
	m := NewMatcher(cat, comp.Track(&trackedPos), comp.Optional(&vel))

	if !m.Seek(e0) {
		t.Fatal("expected Seek to find e0")
	}
	if !vel.Present(&m.Cursor) || vel.At(&m.Cursor).VX != 1 {
		t.Fatalf("expected e0 Vel present with VX=1, got present=%v", vel.Present(&m.Cursor))
	}

	if !m.SeekH(e1) {
		t.Fatal("expected SeekH to report matching archetype for e1")
	}
	if !vel.Present(&m.Cursor) || vel.At(&m.Cursor).VX != 2 {
		t.Fatalf("expected e1 Vel present with VX=2, got present=%v", vel.Present(&m.Cursor))
	}
}
