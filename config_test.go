package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v2"
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
	factory := ecs.NewFactory(&pos)
	factory.Create(10)
	total := 0
	for factory.Next() {
		total += len(factory.IDs)
	}
	if total != 10 {
		t.Errorf("expected 10 entities created, got %d", total)
	}
}
