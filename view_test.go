package goke_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestView_WithTag_And_Without_Logic(t *testing.T) {
	ecs := goke.New()
	// 1. Setup Entities with different structural profiles

	_ = goke.RegisterComponent[position](ecs)
	_ = goke.RegisterComponent[velocity](ecs)
	_ = goke.RegisterComponent[complexComponent](ecs)

	// Entity A: Only position
	blueprintA := goke.NewBlueprint1[position](ecs)
	eA, _ := blueprintA.Create()

	// Entity B: position + velocity (Moving entity)
	blueprintB := goke.NewBlueprint2[position, velocity](ecs)
	eB, _, _ := blueprintB.Create()

	// Entity C: position + complexComponent (Static named entity)
	blueprintC := goke.NewBlueprint2[position, complexComponent](ecs)
	eC, _, _ := blueprintC.Create()

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		view := goke.NewView0(ecs,
			goke.Include[position](),
			goke.Exclude[velocity](),
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
		view := goke.NewView0(ecs,
			goke.Include[position](),
			goke.Include[complexComponent](),
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
		view := goke.NewView0(ecs, goke.Exclude[complexComponent]())

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
