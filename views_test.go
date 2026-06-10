package goke_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
	"github.com/stretchr/testify/assert"
)

func TestView_WithTag_And_Without_Logic(t *testing.T) {
	ecs := goke.New()
	// 1. Setup Entities with different structural profiles

	_ = goke.RegisterComponent[position](ecs)
	_ = goke.RegisterComponent[velocity](ecs)
	_ = goke.RegisterComponent[complexComponent](ecs)

	// Entity A: Only position
	var eA goke.Entity
	blueprintA := goke.NewBlueprint1[position](ecs)
	for page := range blueprintA.Create(1) {
		eA = page.Entity[0]
	}

	// Entity B: position + velocity (Moving entity)
	var eB goke.Entity
	blueprintB := goke.NewBlueprint2[position, velocity](ecs)
	for page := range blueprintB.Create(1) {
		eB = page.Entity[0]
	}

	// Entity C: position + complexComponent (Static named entity)
	var eC goke.Entity
	blueprintC := goke.NewBlueprint2[position, complexComponent](ecs)
	for page := range blueprintC.Create(1) {
		eC = page.Entity[0]
	}

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		view := goke.NewView0(ecs,
			goke.Include[position](),
			goke.Exclude[velocity](),
		)

		found := make(map[uid.UID64]bool)
		for page := range view.All() {
			for _, entity := range page.Entity {
				found[entity] = true
			}
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
		for page := range view.All() {
			for _, entity := range page.Entity {
				assert.Equal(t, eC, entity, "Only Entity C has both position and complexComponent")
				count++
			}
		}
		assert.Equal(t, 1, count)
	})

	// 4. Test: Filter method on a manual slice
	t.Run("Manual Slice Filtering", func(t *testing.T) {
		// Goal: From a list of entities, filter out those that are 'complexComponent'
		view := goke.NewView0(ecs, goke.Exclude[complexComponent]())

		input := []uid.UID64{eA, eB, eC}
		var result []uid.UID64
		for _, head := range view.Filter(input) {
			result = append(result, head.Entity)
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
