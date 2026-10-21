package query

import (
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/iter"
	"github.com/kjkrol/uid"
)

func TestSeek_FindsEntity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	posOpt := comp.Track(&pos)
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, posOpt)
	f := em.CreateFactory(accessSpec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 1, Y: 2}

	m := walk(NewMatcher(cat, posOpt))

	if !m.Seek(e) {
		t.Fatal("expected Seek to find the entity")
	}
	if got := pos.At(&m.Cursor); got.X != 1 || got.Y != 2 {
		t.Errorf("expected {1,2}, got %+v", *got)
	}
}

func TestSeek_UnknownEntity(t *testing.T) {
	cat, _, _ := newQueryCatalog()
	m := walk(NewMatcher(cat))

	if m.Seek(uid.UID64(999)) {
		t.Error("expected Seek to fail for an unknown entity")
	}
}

// Seek caches the table/offsets per archetype. Alternating Seeks between two
// archetypes must re-resolve them each switch, never returning stale data.
func TestSeek_AcrossArchetypes(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var posA iter.ArrayRef[iterPos]
	specA := comp.AccessSpec{}
	specA.Init(cc, comp.Track(&posA))
	fa := em.CreateFactory(specA)
	fa.Create(1)
	fa.Next()
	eA := fa.IDs[0]
	*posA.At(&fa.Cursor) = iterPos{X: 1}

	var posB iter.ArrayRef[iterPos]
	var velB iter.ArrayRef[iterVel]
	specB := comp.AccessSpec{}
	specB.Init(cc, comp.Track(&posB), comp.Track(&velB))
	fb := em.CreateFactory(specB)
	fb.Create(1)
	fb.Next()
	eB := fb.IDs[0]
	*posB.At(&fb.Cursor) = iterPos{X: 2}

	var pos iter.ArrayRef[iterPos]
	m := walk(NewMatcher(cat, comp.Track(&pos)))

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

	var pos iter.ArrayRef[iterPos]
	var vel iter.ArrayRef[iterVel]
	spec := comp.AccessSpec{}
	spec.Init(cc, comp.Track(&pos), comp.Track(&vel))
	f := em.CreateFactory(spec)
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 9}

	// Matcher tracks Pos but excludes Vel — All/Pick would skip this entity,
	// but Seek must still find it since it bypasses the mask entirely.
	m := walk(NewMatcher(cat, comp.Track(&pos), comp.Exclude[iterVel]()))

	if !m.Seek(e) {
		t.Fatal("expected Seek to find the entity regardless of the exclude mask")
	}
	if got := pos.At(&m.Cursor); got.X != 9 {
		t.Errorf("expected X=9, got %v", got.X)
	}
}

// SeekH skips the archetype-change check, trusting the caller to have
// already established the target archetype via a prior Seek on the same
// batch. Within one archetype it must behave exactly like Seek and report a
// match.
func TestSeekH_SameArchetypeMatches(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	posOpt := comp.Track(&pos)
	var accessSpec comp.AccessSpec
	accessSpec.Init(cc, posOpt)
	f := em.CreateFactory(accessSpec)
	f.Create(2)
	f.Next()
	e0, e1 := f.IDs[0], f.IDs[1]
	*pos.At(&f.Cursor) = iterPos{X: 1}
	f.Cursor.Slot = 1
	*pos.At(&f.Cursor) = iterPos{X: 2}

	m := walk(NewMatcher(cat, posOpt))
	if !m.Seek(e0) {
		t.Fatal("expected Seek to find e0")
	}
	if got := pos.At(&m.Cursor); got.X != 1 {
		t.Errorf("expected e0.X == 1, got %v", got.X)
	}

	if !m.SeekH(e1) {
		t.Error("expected SeekH to report a matching archetype for e1")
	}
	if got := pos.At(&m.Cursor); got.X != 2 {
		t.Errorf("expected e1.X == 2, got %v", got.X)
	}
}

// SeekH must report (not silently ignore) an archetype mismatch so the
// caller knows to fall back to Seek.
func TestSeekH_DifferentArchetypeReportsMismatch(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var posA iter.ArrayRef[iterPos]
	specA := comp.AccessSpec{}
	specA.Init(cc, comp.Track(&posA))
	fa := em.CreateFactory(specA)
	fa.Create(1)
	fa.Next()
	eA := fa.IDs[0]
	*posA.At(&fa.Cursor) = iterPos{X: 1}

	var posB iter.ArrayRef[iterPos]
	var velB iter.ArrayRef[iterVel]
	specB := comp.AccessSpec{}
	specB.Init(cc, comp.Track(&posB), comp.Track(&velB))
	fb := em.CreateFactory(specB)
	fb.Create(1)
	fb.Next()
	eB := fb.IDs[0]

	var pos iter.ArrayRef[iterPos]
	m := walk(NewMatcher(cat, comp.Track(&pos)))

	if !m.Seek(eA) {
		t.Fatal("expected Seek to find eA")
	}
	if m.SeekH(eB) {
		t.Error("expected SeekH to report a mismatch for an entity from a different archetype")
	}
	if !m.Seek(eB) {
		t.Error("expected the suggested fallback Seek(eB) to succeed")
	}
}

// Seek bypasses the include/exclude filters but not the column check: an
// archetype missing a tracked component must be refused, cursor untouched.
func TestSeek_RefusesArchetypeMissingATrackedColumn(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	var pos iter.ArrayRef[iterPos]
	var vel iter.ArrayRef[iterVel]
	posOpt, velOpt := comp.Track(&pos), comp.Track(&vel)

	var withBoth comp.AccessSpec
	withBoth.Init(cc, posOpt, velOpt)
	f := em.CreateFactory(withBoth)
	f.Create(1)
	f.Next()
	full := f.IDs[0]
	*pos.At(&f.Cursor) = iterPos{X: 9}

	var posOnly comp.AccessSpec
	posOnly.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f2 := em.CreateFactory(posOnly)
	f2.Create(1)
	f2.Next()
	partial := f2.IDs[0]

	m := walk(NewMatcher(cat, posOpt, velOpt))
	if !m.Seek(full) {
		t.Fatal("expected Seek to find the entity carrying both components")
	}
	if m.Seek(partial) {
		t.Error("Seek returned true for an archetype with no Velocity column, want false")
	}
	if got := pos.At(&m.Cursor); got.X != 9 {
		t.Errorf("cursor moved on a refused Seek: X = %v, want 9", got.X)
	}
}

type seekPad[T any] struct{ V T }

// spawnPosAnd is an entity carrying an iterPos at x and a T, in the archetype of the two
func spawnPosAnd[T any](cc *comp.DefIndex, em *ent.Manager, x float32) uid.UID64 {
	var pos iter.ArrayRef[iterPos]
	var other iter.ArrayRef[T]
	var spec comp.AccessSpec
	spec.Init(cc, comp.Track(&pos), comp.Track(&other))
	f := em.CreateFactory(spec)
	f.Create(1)
	f.Next()
	*pos.At(&f.Cursor) = iterPos{X: x}
	return f.IDs[0]
}

// Seek reads the tables baked as archetypes came: an unfiltered query keeps no tables apart, and
// seeking across its archetypes adds none.
func TestSeek_ReadsTheTablesBakedAsArchetypesCame(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	first := spawnPosAnd[iterVel](cc, em, 1)
	spawnPosAnd[seekPad[int8]](cc, em, 2)
	spawnPosAnd[seekPad[int16]](cc, em, 3)
	spawnPosAnd[seekPad[int32]](cc, em, 4)
	last := spawnPosAnd[seekPad[int64]](cc, em, 5)
	var pos iter.ArrayRef[iterPos]
	m := walk(NewMatcher(cat, comp.Track(&pos)))

	if m.Matcher().seekable != nil {
		t.Fatal("an unfiltered query keeps Seek tables apart")
	}
	for _, step := range []struct {
		id uid.UID64
		x  float32
	}{{last, 5}, {first, 1}, {last, 5}, {first, 1}} {
		if !m.Seek(step.id) || pos.At(&m.Cursor).X != step.x {
			t.Fatalf("Seek(%v): want X %v", step.id, step.x)
		}
	}
	if n := len(m.Matcher().BakedTables); n != 5 {
		t.Errorf("%d tables after seeking, want the 5 archetypes", n)
	}
}

// A filtered query bakes apart, for Seek, the archetypes holding every tracked component, the
// filtered out among them; one lacking a tracked component stays out.
func TestSeek_FilteredQueryBakesApartTheArchetypesWithItsColumns(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	kept := spawnPosAnd[seekPad[int8]](cc, em, 1)
	excluded := spawnPosAnd[iterVel](cc, em, 2)
	var velOnly iter.ArrayRef[iterVel]
	var spec comp.AccessSpec
	spec.Init(cc, comp.Track(&velOnly))
	f := em.CreateFactory(spec)
	f.Create(1)
	f.Next()
	without := f.IDs[0]
	var pos iter.ArrayRef[iterPos]
	m := walk(NewMatcher(cat, comp.Track(&pos), comp.Exclude[iterVel]()))

	if n := len(m.Matcher().BakedTables); n != 1 {
		t.Errorf("%d matched tables, want 1", n)
	}
	if s := m.Matcher().seekable; s == nil || len(s.BakedTables) != 2 {
		t.Fatalf("Seek tables apart: %v, want the 2 archetypes holding iterPos", s)
	}
	if !m.Seek(kept) || pos.At(&m.Cursor).X != 1 {
		t.Error("Seek into the matched archetype")
	}
	if !m.Seek(excluded) || pos.At(&m.Cursor).X != 2 {
		t.Error("Seek into the archetype the filter leaves out")
	}
	if m.Seek(without) {
		t.Error("Seek into an archetype without iterPos")
	}
}
