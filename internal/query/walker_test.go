package query

import (
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
)

// walk is a Walker over m, standing nowhere
func walk(m *Matcher) *Walker {
	var w Walker
	w.Init(m)
	return &w
}

// Clear ends the walk, not what the Matcher matches: Next reports nothing till a new walk starts.
func TestWalker_ClearEndsTheWalkOnly(t *testing.T) {
	cat, cc, em := newQueryCatalog()
	var spec comp.AccessSpec
	spec.Init(cc, comp.Track(new(iter.ArrayRef[iterPos])))
	f := em.CreateFactory(spec)
	f.Create(1)
	f.Next()
	w := walk(NewMatcher(cat, comp.Track(new(iter.ArrayRef[iterPos]))))

	if !w.All().Next() {
		t.Fatal("the walk found nothing")
	}
	w.Clear()
	if w.Next() {
		t.Error("Next after Clear goes on walking")
	}
	if len(w.Matcher().BakedTables) != 1 || !w.All().Next() {
		t.Error("Clear changed what the Matcher matches")
	}
}
