package reg_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

func TestRegistry_AddMatcher_MatchesExistingArchetype(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	var pos iter.Col[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(1)
	factory.Next()
	pos.Slice(&factory.Cursor)[0] = Position{X: 1, Y: 2}

	var trackedPos iter.Col[Position]
	matcher := r.AddMatcher(comp.Track(&trackedPos))

	count := 0
	matcher.All()
	for matcher.Next() {
		for range matcher.Cursor.IDs {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected matcher to find the already-created entity, got %d", count)
	}
}

// Registry.Init wires MatcherCatalog.OnArchetypeCreated to EntityManager —
// a Matcher added before an archetype exists must still pick it up once
// entities are created into it. This is Registry's own responsibility: it
// glues two otherwise-independent subsystems (ent.Manager, query.Catalog)
// together.
func TestRegistry_AddMatcher_ReactsToArchetypeCreatedAfterward(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	var trackedPos iter.Col[Position]
	matcher := r.AddMatcher(comp.Track(&trackedPos))

	count := 0
	matcher.All()
	for matcher.Next() {
		for range matcher.Cursor.IDs {
			count++
		}
	}
	if count != 0 {
		t.Fatalf("expected no matches before any entity exists, got %d", count)
	}

	var pos iter.Col[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(1)
	factory.Next()

	count = 0
	matcher.All()
	for matcher.Next() {
		for range matcher.Cursor.IDs {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected the matcher to react to the newly created archetype, got %d", count)
	}
}

func TestRegistry_AddMatcher_RespectsExclude(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())
	r.RegComp(reflect.TypeFor[Tag]())

	var pos iter.Col[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(1)
	factory.Next()

	matcher := r.AddMatcher(comp.Include[Position](), comp.Exclude[Tag]())

	count := 0
	matcher.All()
	for matcher.Next() {
		for range matcher.Cursor.IDs {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 match (Position without Tag), got %d", count)
	}
}
