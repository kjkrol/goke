package query

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
	"github.com/kjkrol/uid"
)

func TestSeek_FindsEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.Col[iterPos]
	posOpt := comp.Track(&pos)
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, posOpt)
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 1, Y: 2}

	m := NewMatcher(cat, posOpt)

	if !m.Seek(e) {
		t.Fatal("expected Seek to find the entity")
	}
	if got := pos.At(&m.Cursor); got.X != 1 || got.Y != 2 {
		t.Errorf("expected {1,2}, got %+v", *got)
	}
}

func TestSeek_UnknownEntity(t *testing.T) {
	cat, _, _ := newQueryCatalog()
	m := NewMatcher(cat)

	if m.Seek(uid.UID64(999)) {
		t.Error("expected Seek to fail for an unknown entity")
	}
}

// Seek caches the table/offsets per archetype. Alternating Seeks between two
// archetypes must re-resolve them each switch, never returning stale data.
func TestSeek_AcrossArchetypes(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var posA iter.Col[iterPos]
	specA := comp.AccessSpec{}
	specA.Init(cc, comp.Track(&posA))
	fa := em.CreateFactory(specA)
	fa.Create(1)
	fa.Next()
	eA := fa.IDs[0]
	*posA.At(&fa.Cursor) = iterPos{X: 1}

	var posB iter.Col[iterPos]
	var velB iter.Col[iterVel]
	specB := comp.AccessSpec{}
	specB.Init(cc, comp.Track(&posB), comp.Track(&velB))
	fb := em.CreateFactory(specB)
	fb.Create(1)
	fb.Next()
	eB := fb.IDs[0]
	*posB.At(&fb.Cursor) = iterPos{X: 2}

	var pos iter.Col[iterPos]
	m := NewMatcher(cat, comp.Track(&pos))

	for i := 0; i < 3; i++ {
		if !m.Seek(eA) || pos.At(&m.Cursor).X != 1 {
			t.Fatalf("iter %d: expected eA.X == 1", i)
		}
		if !m.Seek(eB) || pos.At(&m.Cursor).X != 2 {
			t.Fatalf("iter %d: expected eB.X == 2", i)
		}
	}
}

// Seek is independent of the Matcher's include/exclude mask: it trusts the
// caller and resolves whatever components are tracked, even if the entity's
// archetype wouldn't itself match the mask via All/Pick.
func TestSeek_IgnoresIncludeExcludeMask(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.Col[iterPos]
	var vel iter.Col[iterVel]
	spec := comp.AccessSpec{}
	spec.Init(cc, comp.Track(&pos), comp.Track(&vel))
	f := em.CreateFactory(spec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 9}

	// Matcher tracks Pos but excludes Vel — All/Pick would skip this entity,
	// but Seek must still find it since it bypasses the mask entirely.
	m := NewMatcher(cat, comp.Track(&pos), comp.Exclude[iterVel]())

	if !m.Seek(e) {
		t.Fatal("expected Seek to find the entity regardless of the exclude mask")
	}
	if got := pos.At(&m.Cursor); got.X != 9 {
		t.Errorf("expected X=9, got %v", got.X)
	}
}
