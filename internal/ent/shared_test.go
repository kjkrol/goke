package ent_test

import (
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
)

// shared_test.go contains component types and small helpers shared across
// this package's test files.

type Position struct{ X, Y float64 }
type Velocity struct{ VX, VY float64 }
type Tag struct{}

func newManager() *ent.Manager {
	var m ent.Manager
	m.Init(ent.DefaultConfig(), nil)
	return &m
}

// tagAccessSpec builds an AccessSpec that anchors entities to a tag-only
// archetype (real, non-root). Root itself is never a valid Factory target:
// Catalog.Init creates it directly without invoking onArchetypeCreated, so
// its table never gets an IDSeeder wired up.
func tagAccessSpec(tagID comp.ID) comp.AccessSpec {
	var spec comp.AccessSpec
	if err := spec.Tag(tagID); err != nil {
		panic(err)
	}
	return spec
}
