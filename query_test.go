package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v2"
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
			for _, entityID := range query.Cursor().IDs {
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
			for _, entityID := range query.Cursor().IDs {
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
			result = append(result, query.Entity())
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
			for _, entityID := range query.Cursor().IDs {
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
			for _, entityID := range query.Cursor().IDs {
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
		count += len(query.Cursor().IDs)
	}
	assert.Equal(t, 3, count)
}

// --- Query method tests (All, Pick, Seek, SeekH, Cursor, Entity, Idx) ---

func TestQuery_All_SlicesCoverAllEntities(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(5)
	for factory.Next() {
		for i := range factory.IDs {
			pos.Slice(&factory.Cursor)[i] = Position{X: float32(i)}
		}
	}

	query := ecs.NewQueryBuilder(&pos).Build()
	sum := float32(0)
	count := 0
	query.All()
	for query.Next() {
		cursor := query.Cursor()
		posSlice := pos.Slice(cursor)
		for i := range cursor.IDs {
			sum += posSlice[i].X
			count++
		}
	}
	assert.Equal(t, 5, count)
	assert.Equal(t, float32(0+1+2+3+4), sum)
}

func TestQuery_Pick_EntityAndIdxMatchInput(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(3)
	factory.Next()
	ids := append([]uid.UID64{}, factory.IDs...)

	query := ecs.NewQueryBuilder(&pos).Build()
	query.Pick(ids)
	i := 0
	for query.Next() {
		assert.Equal(t, ids[i], query.Entity(), "Entity() should match the input slice at Idx()")
		assert.Equal(t, i, query.Idx())
		i++
	}
	assert.Equal(t, 3, i)
}

func TestQuery_Seek_FindsEntityAndFailsForMissing(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]
	pos.At(&factory.Cursor).X = 42

	query := ecs.NewQueryBuilder(&pos).Build()
	if !query.Seek(entityID) {
		t.Fatalf("expected Seek to find the entity")
	}
	assert.Equal(t, float32(42), pos.At(query.Cursor()).X)

	if query.Seek(uid.UID64(999999)) {
		t.Errorf("expected Seek to return false for a nonexistent entity")
	}
}

func TestQuery_SeekH_SameArchetypeMatches(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(2)
	factory.Next()
	pos.Slice(&factory.Cursor)[0] = Position{X: 1}
	pos.Slice(&factory.Cursor)[1] = Position{X: 2}
	e0, e1 := factory.IDs[0], factory.IDs[1]

	query := ecs.NewQueryBuilder(&pos).Build()
	if !query.Seek(e0) {
		t.Fatalf("expected Seek to find e0")
	}
	assert.Equal(t, float32(1), pos.At(query.Cursor()).X)

	// e1 is alive and shares e0's archetype (both spawned from the same
	// factory batch), so SeekH should report a match and position correctly.
	if !query.SeekH(e1) {
		t.Errorf("expected SeekH to report a matching archetype")
	}
	assert.Equal(t, float32(2), pos.At(query.Cursor()).X)
}

func TestQuery_SeekH_DifferentArchetypeReportsMismatch(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]

	// e0: Position only.
	factory0 := ecs.NewFactory(&pos)
	factory0.Create(1)
	factory0.Next()
	e0 := factory0.IDs[0]

	// e1: Position + Velocity — a different archetype from e0.
	factory1 := ecs.NewFactory(&pos, &vel)
	factory1.Create(1)
	factory1.Next()
	e1 := factory1.IDs[0]

	query := ecs.NewQueryBuilder(&pos).Build()
	if !query.Seek(e0) {
		t.Fatalf("expected Seek to find e0")
	}

	// SeekH must report the mismatch instead of silently using e0's cached
	// table — the caller is expected to fall back to Seek(e1) instead.
	if query.SeekH(e1) {
		t.Errorf("expected SeekH to report a mismatch for an entity from a different archetype")
	}
	if !query.Seek(e1) {
		t.Errorf("expected the suggested fallback Seek(e1) to succeed")
	}
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
