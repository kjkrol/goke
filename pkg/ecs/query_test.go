package ecs_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/pkg/ecs"
	"github.com/stretchr/testify/assert"
)

func TestQuery_WithTag_And_Without_Logic(t *testing.T) {
	engine := ecs.NewEngine()

	// 1. Setup Entities with different structural profiles

	positionType := engine.RegisterComponentType(reflect.TypeFor[position]())
	velocityType := engine.RegisterComponentType(reflect.TypeFor[velocity]())
	complexType := engine.RegisterComponentType(reflect.TypeFor[complexComponent]())

	// Entity A: Only position
	eA := engine.CreateEntity()
	engine.Allocate(eA, positionType)

	// Entity B: position + velocity (Moving entity)
	eB := engine.CreateEntity()
	engine.Allocate(eB, positionType)
	engine.Allocate(eB, velocityType)

	// Entity C: position + complexComponent (Static named entity)
	eC := engine.CreateEntity()
	engine.Allocate(eC, positionType)
	engine.Allocate(eC, complexType)

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		query := ecs.NewQuery0(engine,
			ecs.WithTag[position](),
			ecs.Without[velocity](),
		)

		found := make(map[ecs.Entity]bool)
		for e := range query.All() {
			found[e] = true
		}

		assert.Len(t, found, 2, "Should find exactly 2 entities (A and C)")
		assert.True(t, found[eA], "Entity A (position only) should match")
		assert.True(t, found[eC], "Entity C (position + complex) should match")
		assert.False(t, found[eB], "Entity B (velocity) should be excluded")
	})

	// 3. Test: Tag as a mandatory requirement
	t.Run("Tag as Requirement", func(t *testing.T) {
		// Goal: Find entities with 'position' AND 'complexComponent'
		// Expected: eC only
		query := ecs.NewQuery0(engine,
			ecs.WithTag[position](),
			ecs.WithTag[complexComponent](),
		)

		count := 0
		for e := range query.All() {
			assert.Equal(t, eC, e, "Only Entity C has both position and complexComponent")
			count++
		}
		assert.Equal(t, 1, count)
	})

	// 4. Test: Filter method on a manual slice
	t.Run("Manual Slice Filtering", func(t *testing.T) {
		// Goal: From a list of entities, filter out those that are 'complexComponent'
		query := ecs.NewQuery0(engine, ecs.Without[complexComponent]())

		input := []ecs.Entity{eA, eB, eC}
		var result []ecs.Entity
		for e := range query.Filter(input) {
			result = append(result, e)
		}

		assert.Len(t, result, 2, "Should skip Entity C")
		assert.Contains(t, result, eA)
		assert.Contains(t, result, eB)
		assert.NotContains(t, result, eC)
	})
}

// Any other shared test utilities can go here, for example:
type complexComponent struct {
	Active bool
	Layer  int32
	Name   [16]byte
}

type position struct {
	x, y float64
}

type velocity struct {
	vx, vy float64
}
