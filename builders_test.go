package goke_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder_IncludeExclude(t *testing.T) {
	ecs := goke.New()
	// 1. Setup Entities with different structural profiles

	_ = goke.RegComp[position](ecs)
	_ = goke.RegComp[velocity](ecs)
	_ = goke.RegComp[complexComponent](ecs)

	// Entity A: Only position
	var eA uid.UID64
	factoryA := ecs.NewFactory(new(goke.Comp[position]))
	factoryA.Create(1)
	factoryA.Next()
	eA = factoryA.IDs[0]

	// Entity B: position + velocity (Moving entity)
	var eB uid.UID64
	factoryB := ecs.NewFactory(new(goke.Comp[position]), new(goke.Comp[velocity]))
	factoryB.Create(1)
	factoryB.Next()
	eB = factoryB.IDs[0]

	// Entity C: position + complexComponent (Static named entity)
	var eC uid.UID64
	factoryC := ecs.NewFactory(new(goke.Comp[position]), new(goke.Comp[complexComponent]))
	factoryC.Create(1)
	factoryC.Next()
	eC = factoryC.IDs[0]

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		query := ecs.NewQueryBuilder().Include(goke.Include[position]()).Exclude(goke.Exclude[velocity]()).Build()

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
		query := ecs.NewQueryBuilder().Include(goke.Include[position](), goke.Include[complexComponent]()).Build()

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
		query := ecs.NewQueryBuilder().Exclude(goke.Exclude[complexComponent]()).Build()

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

	// 5. Test: Include/Exclude accumulate across separate chained calls, not
	// just across multiple opts within a single call.
	t.Run("Chained Include and Exclude calls accumulate", func(t *testing.T) {
		query := ecs.NewQueryBuilder().
			Include(goke.Include[position]()).
			Include(goke.Include[complexComponent]()).
			Exclude(goke.Exclude[velocity]()).
			Build()

		count := 0
		query.All()
		for query.Next() {
			for _, entityID := range query.Cursor.IDs {
				assert.Equal(t, eC, entityID)
				count++
			}
		}
		assert.Equal(t, 1, count, "expected only eC (position+complexComponent, no velocity)")
	})

	// 6. Test: a trackable handle passed to NewQueryBuilder combines correctly
	// with subsequent Include/Exclude opts.
	t.Run("Trackable combined with Include/Exclude", func(t *testing.T) {
		var pos position
		_ = pos
		var posComp goke.Comp[position]
		query := ecs.NewQueryBuilder(&posComp).Exclude(goke.Exclude[velocity]()).Build()

		found := make(map[uid.UID64]bool)
		query.All()
		for query.Next() {
			for _, entityID := range query.Cursor.IDs {
				found[entityID] = true
			}
		}
		assert.Len(t, found, 2, "expected eA and eC (have position, lack velocity)")
		assert.True(t, found[eA])
		assert.True(t, found[eC])
	})
}

func TestQueryBuilder_EmptyBuild(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	factory := ecs.NewFactory(new(goke.Comp[Position]))
	factory.Create(3)
	factory.Next()

	// A builder with no handles and no Include/Exclude opts matches every
	// entity — an empty mask is not a constraint.
	query := ecs.NewQueryBuilder().Build()
	count := 0
	query.All()
	for query.Next() {
		count += len(query.Cursor.IDs)
	}
	assert.Equal(t, 3, count)
}

func TestEditorBuilder_AddComp(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	// Entity starts with only Velocity.
	var vel goke.Comp[Velocity]
	factory := ecs.NewFactory(&vel)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	// Add Position and write its value through the editor's cursor.
	var pos goke.Comp[Position]
	editor := ecs.NewEditorBuilder(&pos).Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}
	pos.At(&editor.Cursor).X = 55

	val := seekComp[Position](ecs, entityID)
	if val == nil || val.X != 55 {
		t.Errorf("expected Position.X == 55, got %v", val)
	}
}

func TestEditorBuilder_InvalidEntity(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Comp[Position]
	editor := ecs.NewEditorBuilder(&pos).Build()

	if editor.Update(uid.UID64(999)) {
		t.Errorf("expected Update to return false for a nonexistent entity")
	}
}

func TestEditorBuilder_Delete(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	factory := ecs.NewFactory(&pos, &vel)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	editor := ecs.NewEditorBuilder().Delete(goke.Del[Velocity]()).Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}

	if hasComp[Velocity](ecs, entityID) {
		t.Errorf("expected Velocity to be removed")
	}
	if p := seekComp[Position](ecs, entityID); p == nil {
		t.Errorf("expected Position to remain")
	}
}

func TestEditorBuilder_AddAndDelete(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var vel goke.Comp[Velocity]
	factory := ecs.NewFactory(&vel)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	// Swap Velocity for Position in a single migration.
	var pos goke.Comp[Position]
	editor := ecs.NewEditorBuilder(&pos).Delete(goke.Del[Velocity]()).Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}
	pos.At(&editor.Cursor).Y = 7

	if hasComp[Velocity](ecs, entityID) {
		t.Errorf("expected Velocity to be removed")
	}
	p := seekComp[Position](ecs, entityID)
	if p == nil || p.Y != 7 {
		t.Errorf("expected Position.Y == 7, got %v", p)
	}
}

func TestEditorBuilder_ChainedDelete(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)
	_ = goke.RegComp[Discount](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var disc goke.Comp[Discount]
	factory := ecs.NewFactory(&pos, &vel, &disc)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	editor := ecs.NewEditorBuilder().
		Delete(goke.Del[Velocity]()).
		Delete(goke.Del[Discount]()).
		Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}

	if hasComp[Velocity](ecs, entityID) {
		t.Errorf("expected Velocity to be removed")
	}
	if hasComp[Discount](ecs, entityID) {
		t.Errorf("expected Discount to be removed")
	}
	if pos := seekComp[Position](ecs, entityID); pos == nil {
		t.Errorf("expected Position to remain")
	}
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

// Shared component types and helpers used by builder tests.
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
