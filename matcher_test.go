package goke_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
	"github.com/stretchr/testify/assert"
)

func TestMatcher_IncludeExclude(t *testing.T) {
	ecs := goke.New()
	// 1. Setup Entities with different structural profiles

	_ = goke.RegComp[position](ecs)
	_ = goke.RegComp[velocity](ecs)
	_ = goke.RegComp[complexComponent](ecs)

	// Entity A: Only position
	var eA uid.UID64
	factoryA := ecs.CreateFactory(goke.Add(new(goke.Col[position])))
	factoryA.Create(1)
	factoryA.Next()
	eA = factoryA.IDs[0]

	// Entity B: position + velocity (Moving entity)
	var eB uid.UID64
	factoryB := ecs.CreateFactory(goke.Add(new(goke.Col[position])), goke.Add(new(goke.Col[velocity])))
	factoryB.Create(1)
	factoryB.Next()
	eB = factoryB.IDs[0]

	// Entity C: position + complexComponent (Static named entity)
	var eC uid.UID64
	factoryC := ecs.CreateFactory(goke.Add(new(goke.Col[position])), goke.Add(new(goke.Col[complexComponent])))
	factoryC.Create(1)
	factoryC.Next()
	eC = factoryC.IDs[0]

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		query := ecs.CreateMatcher(goke.Include[position](),
			goke.Exclude[velocity](),
		)

		found := make(map[uid.UID64]bool)
		query.All()
		for query.Next() {
			for _, entityID := range query.Cursor.IDs {
				found[entityID] = true
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
		query := ecs.CreateMatcher(goke.Include[position](),
			goke.Include[complexComponent](),
		)

		count := 0
		query.All()
		for query.Next() {
			for _, entityID := range query.Cursor.IDs {
				assert.Equal(t, eC, entityID, "Only Entity C has both position and complexComponent")
				count++
			}
		}
		assert.Equal(t, 1, count)
	})

	// 4. Test: Pick method on a manual slice
	t.Run("Manual Slice Filtering", func(t *testing.T) {
		// Goal: From a list of entities, filter out those that are 'complexComponent'
		query := ecs.CreateMatcher(goke.Exclude[complexComponent]())

		input := []uid.UID64{eA, eB, eC}
		var result []uid.UID64
		query.Pick(input)
		for query.Next() {
			result = append(result, query.Entity)
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
