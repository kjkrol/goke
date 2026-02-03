package goke_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestView_WithTag_And_Without_Logic(t *testing.T) {
	eng := goke.NewEngine()
	// 1. Setup Entities with different structural profiles

	positionType := goke.ComponentRegister[position](eng)
	velocityType := goke.ComponentRegister[velocity](eng)
	complexType := goke.ComponentRegister[complexComponent](eng)

	// Entity A: Only position
	eA := goke.EntityCreate(eng)
	goke.EntityAllocateComponentByInfo[position](eng, eA, positionType)

	// Entity B: position + velocity (Moving entity)
	eB := goke.EntityCreate(eng)
	goke.EntityAllocateComponentByInfo[position](eng, eB, positionType)
	goke.EntityAllocateComponentByInfo[velocity](eng, eB, velocityType)

	// Entity C: position + complexComponent (Static named entity)
	eC := goke.EntityCreate(eng)
	goke.EntityAllocateComponentByInfo[position](eng, eC, positionType)
	goke.EntityAllocateComponentByInfo[complexComponent](eng, eC, complexType)

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		view := goke.NewView0(eng,
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
		view := goke.NewView0(eng,
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
		view := goke.NewView0(eng, core.Without[complexComponent]())

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
