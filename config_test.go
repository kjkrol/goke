package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v3"
)

func TestWithEntityCap(t *testing.T) {
	var c goke.Config
	goke.WithEntityCap(123)(&c)

	if c.Entity.Cap != 123 {
		t.Errorf("expected Entity.Cap 123, got %d", c.Entity.Cap)
	}
}

func TestWithEntityFreeCap(t *testing.T) {
	var c goke.Config
	goke.WithEntityFreeCap(456)(&c)

	if c.Entity.FreeCap != 456 {
		t.Errorf("expected Entity.FreeCap 456, got %d", c.Entity.FreeCap)
	}
}

func TestECSOptions_AppliedByNew(t *testing.T) {
	ecs := goke.New(goke.WithEntityCap(10), goke.WithEntityFreeCap(20))
	if ecs == nil {
		t.Fatal("expected a non-nil ECS")
	}

	// The options must actually take effect, not just be accepted silently:
	// an entity pool with Cap=10 should still be usable for creating entities.
	var pos goke.Comp[Position]
	total := 0
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(10)
		for factory.Next() {
			total += len(factory.IDs)
		}
	}})
	if total != 10 {
		t.Errorf("expected 10 entities created, got %d", total)
	}
}

func TestWithMatcherCap(t *testing.T) {
	var c goke.Config
	goke.WithMatcherCap(789)(&c)

	if c.Matcher.Cap != 789 {
		t.Errorf("expected Matcher.Cap 789, got %d", c.Matcher.Cap)
	}
}

// A world builds as many queries as it needs: past the catalog's initial room, the first query
// and the last both see the entities spawned after them.
func TestECS_BuildsMoreQueriesThanItHadRoomFor(t *testing.T) {
	type Pos struct{ X, Y float32 }
	ecs := goke.New(goke.WithMatcherCap(4))
	var queries []*goke.Query
	var pos goke.Comp[Pos]
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		for range 200 {
			queries = append(queries, si.NewQueryBuilder(&pos).Build())
		}
		f := si.NewFactory(&pos)
		f.Create(3)
		for f.Next() {
		}
	}})
	for _, i := range []int{0, 3, 4, 199} {
		n := 0
		for q := queries[i].All(); q.Next(); {
			n += len(q.Cursor().IDs)
		}
		if n != 3 {
			t.Errorf("query %d of 200 sees %d entities, want 3", i, n)
		}
	}
}
