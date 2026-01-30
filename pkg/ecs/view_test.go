package ecs_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/pkg/ecs"
	"github.com/stretchr/testify/assert"
)

func TestView_WithTag_And_Without_Logic(t *testing.T) {
	eng := ecs.NewEngine()
	// 1. Setup Entities with different structural profiles

	positionType := eng.RegisterComponentType(reflect.TypeFor[position]())
	velocityType := eng.RegisterComponentType(reflect.TypeFor[velocity]())
	complexType := eng.RegisterComponentType(reflect.TypeFor[complexComponent]())

	// Entity A: Only position
	eA := eng.CreateEntity()
	ecs.AllocateComponentByInfo[position](eng, eA, positionType)

	// Entity B: position + velocity (Moving entity)
	eB := eng.CreateEntity()
	ecs.AllocateComponentByInfo[position](eng, eB, positionType)
	ecs.AllocateComponentByInfo[velocity](eng, eB, velocityType)

	// Entity C: position + complexComponent (Static named entity)
	eC := eng.CreateEntity()
	ecs.AllocateComponentByInfo[position](eng, eC, positionType)
	ecs.AllocateComponentByInfo[complexComponent](eng, eC, complexType)

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		view := ecs.NewView0(eng,
			core.WithTag[position](),
			core.Without[velocity](),
		)

		found := make(map[core.Entity]bool)
		for e := range view.All() {
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
		view := ecs.NewView0(eng,
			core.WithTag[position](),
			core.WithTag[complexComponent](),
		)

		count := 0
		for e := range view.All() {
			assert.Equal(t, eC, e, "Only Entity C has both position and complexComponent")
			count++
		}
		assert.Equal(t, 1, count)
	})

	// 4. Test: Filter method on a manual slice
	t.Run("Manual Slice Filtering", func(t *testing.T) {
		// Goal: From a list of entities, filter out those that are 'complexComponent'
		view := ecs.NewView0(eng, core.Without[complexComponent]())

		input := []core.Entity{eA, eB, eC}
		var result []core.Entity
		for e := range view.Filter(input) {
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
