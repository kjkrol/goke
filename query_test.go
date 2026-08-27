package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder_IncludeExclude(t *testing.T) {
	ecs := goke.New()
	// 1. Setup Entities with different structural profiles

	_ = ecs.RegComp[position]()
	_ = ecs.RegComp[velocity]()
	_ = ecs.RegComp[complexComponent]()

	// Entity A: Only position
	var eA uid.UID64
	// Entity B: position + velocity (Moving entity)
	var eB uid.UID64
	// Entity C: position + complexComponent (Static named entity)
	var eC uid.UID64
	var posComp goke.Comp[position]
	var inclExclQuery, tagReqQuery, sliceFilterQuery, chainedQuery, trackableQuery *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factoryA := si.NewFactory(new(goke.Comp[position]))
		factoryA.Create(1)
		factoryA.Next()
		eA = factoryA.IDs[0]

		factoryB := si.NewFactory(new(goke.Comp[position]), new(goke.Comp[velocity]))
		factoryB.Create(1)
		factoryB.Next()
		eB = factoryB.IDs[0]

		factoryC := si.NewFactory(new(goke.Comp[position]), new(goke.Comp[complexComponent]))
		factoryC.Create(1)
		factoryC.Next()
		eC = factoryC.IDs[0]

		inclExclQuery = si.NewQueryBuilder().Include(goke.Include[position]()).Exclude(goke.Exclude[velocity]()).Build()
		tagReqQuery = si.NewQueryBuilder().Include(goke.Include[position](), goke.Include[complexComponent]()).Build()
		sliceFilterQuery = si.NewQueryBuilder().Exclude(goke.Exclude[complexComponent]()).Build()
		chainedQuery = si.NewQueryBuilder().
			Include(goke.Include[position]()).
			Include(goke.Include[complexComponent]()).
			Exclude(goke.Exclude[velocity]()).
			Build()
		trackableQuery = si.NewQueryBuilder(&posComp).Exclude(goke.Exclude[velocity]()).Build()
	}})

	// 2. Test: Filter Inclusion (WithTag) and Exclusion (Without)
	t.Run("Inclusion and Exclusion Logic", func(t *testing.T) {
		// Goal: Find entities that have 'position', but are NOT 'velocity' (not moving)
		// Expected: eA and eC
		query := inclExclQuery

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
		query := tagReqQuery

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
		query := sliceFilterQuery

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
		query := chainedQuery

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
		query := trackableQuery

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
	_ = ecs.RegComp[Position]()

	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(new(goke.Comp[Position]))
		factory.Create(3)
		factory.Next()

		// A builder with no handles and no Include/Exclude opts matches every
		// entity — an empty mask is not a constraint.
		query = si.NewQueryBuilder().Build()
	}})
	count := 0
	query.All()
	for query.Next() {
		count += len(query.Cursor().IDs)
	}
	assert.Equal(t, 3, count)
}

func TestQueryBuilder_Optional(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[position]()
	_ = ecs.RegComp[velocity]()

	// eA: position only — velocity absent. eB: position + velocity — present.
	var eA, eB uid.UID64
	var velComp goke.OptComp[velocity]
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factoryA := si.NewFactory(new(goke.Comp[position]))
		factoryA.Create(1)
		factoryA.Next()
		eA = factoryA.IDs[0]

		var vel goke.Comp[velocity]
		factoryB := si.NewFactory(new(goke.Comp[position]), &vel)
		factoryB.Create(1)
		factoryB.Next()
		eB = factoryB.IDs[0]
		*vel.At(&factoryB.Cursor) = velocity{VX: 9, VY: 9}

		query = si.NewQueryBuilder(new(goke.Comp[position])).Optional(&velComp).Build()
	}})

	seen := map[uid.UID64]bool{}
	query.All()
	for query.Next() {
		cur := query.Cursor()
		present := velComp.Present(cur)
		for i, e := range cur.IDs {
			seen[e] = true
			switch e {
			case eA:
				assert.False(t, present, "eA's chunk should report velocity absent")
				assert.Nil(t, velComp.Slice(cur), "expected nil Slice for eA's chunk")
			case eB:
				assert.True(t, present, "eB's chunk should report velocity present")
				assert.Equal(t, velocity{VX: 9, VY: 9}, velComp.Slice(cur)[i])
			}
		}
	}

	// Optional never gates the match — both entities are matched by a query
	// that only requires position.
	assert.True(t, seen[eA])
	assert.True(t, seen[eB])

	// OptComp.At (Pick-mode) mirrors Slice's present/absent behavior.
	query.Pick([]uid.UID64{eA, eB})
	for query.Next() {
		cur := query.Cursor()
		switch query.Entity() {
		case eA:
			assert.Nil(t, velComp.At(cur), "expected nil At for eA (velocity absent)")
		case eB:
			assert.Equal(t, velocity{VX: 9, VY: 9}, *velComp.At(cur))
		}
	}
}

// --- Query method tests (All, Pick, Seek, SeekH, Cursor, Entity, Idx) ---

func TestQuery_All_SlicesCoverAllEntities(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()

	var pos goke.Comp[Position]
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(5)
		for factory.Next() {
			for i := range factory.IDs {
				pos.Slice(&factory.Cursor)[i] = Position{X: float32(i)}
			}
		}
		query = si.NewQueryBuilder(&pos).Build()
	}})

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
	_ = ecs.RegComp[Position]()

	var pos goke.Comp[Position]
	var ids []uid.UID64
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(3)
		factory.Next()
		ids = append([]uid.UID64{}, factory.IDs...)

		query = si.NewQueryBuilder(&pos).Build()
	}})

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
	_ = ecs.RegComp[Position]()

	var pos goke.Comp[Position]
	var entityID uid.UID64
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(1)
		factory.Next()
		entityID = factory.IDs[0]
		pos.At(&factory.Cursor).X = 42

		query = si.NewQueryBuilder(&pos).Build()
	}})

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
	_ = ecs.RegComp[Position]()

	var pos goke.Comp[Position]
	var e0, e1 uid.UID64
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(2)
		factory.Next()
		pos.Slice(&factory.Cursor)[0] = Position{X: 1}
		pos.Slice(&factory.Cursor)[1] = Position{X: 2}
		e0, e1 = factory.IDs[0], factory.IDs[1]

		query = si.NewQueryBuilder(&pos).Build()
	}})

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
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Velocity]()

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var e0, e1 uid.UID64
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		// e0: Position only.
		factory0 := si.NewFactory(&pos)
		factory0.Create(1)
		factory0.Next()
		e0 = factory0.IDs[0]

		// e1: Position + Velocity — a different archetype from e0.
		factory1 := si.NewFactory(&pos, &vel)
		factory1.Create(1)
		factory1.Next()
		e1 = factory1.IDs[0]

		query = si.NewQueryBuilder(&pos).Build()
	}})

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

func TestQuery_Clear_NextYieldsNothing(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()

	var pos goke.Comp[Position]
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(1)
		factory.Next()
		query = si.NewQueryBuilder(&pos).Build()
	}})

	query.All()
	if !query.Next() {
		t.Fatalf("expected Next to find the spawned entity before Clear")
	}

	query.Clear()

	query.All()
	if query.Next() {
		t.Error("expected Next to yield nothing after Clear")
	}
}

// Shared component types and helpers used by builder tests.
type complexComponent struct {
	Active bool
	Layer  int32
	Name   [16]byte
}

type position struct {
	X, Y float64
}

type velocity struct {
	VX, VY float64
}
