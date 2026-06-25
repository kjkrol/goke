package goke_test

import (
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// shared_test.go contains component types and small helpers shared across
// this package's test files.

type Position struct{ X, Y float32 }
type Velocity struct{ VX, VY float32 }
type Discount struct{ Percentage float64 }

// seekComp reads a single entity's component via Query.Seek.
// Returns nil if the entity does not exist.
func seekComp[T any](ecs *goke.ECS, e uid.UID64) *T {
	var col goke.Comp[T]
	m := ecs.NewQueryBuilder(&col).Build()
	if !m.Seek(e) {
		return nil
	}
	return col.At(m.Cursor())
}

// hasComp reports whether e currently has component T, checked via the
// Query's include mask (Pick) rather than Seek — Seek is mask-independent
// and would happily resolve garbage offsets for a component the entity no
// longer has, so it cannot be used to detect absence on an entity that still
// has other components.
func hasComp[T any](ecs *goke.ECS, e uid.UID64) bool {
	q := ecs.NewQueryBuilder().Include(goke.Include[T]()).Build()
	q.Pick([]uid.UID64{e})
	return q.Next()
}
