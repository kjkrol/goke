package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

func TestRemover_RemovesMatchingEntities(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()

	var pos goke.Comp[Position]
	var ids []uid.UID64
	var query *goke.Query
	var remover *goke.Remover
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(3)
		for factory.Next() {
			ids = append(ids, factory.IDs...)
		}

		query = si.NewQueryBuilder(&pos).Build()
		remover = si.Remover()
	}})

	sys := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			buf := query.BeginMigrate(cb)
			for _, id := range cursor.IDs {
				buf.Add(id)
			}
			buf.Commit(remover)
		}
	}})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if hasComp(query, id) {
			t.Errorf("entity %v: expected already removed via cb.Commit(remover), but still matches", id)
		}
	}
}

func TestRemover_LeavesNonMatchingEntitiesIntact(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Velocity]()

	var pos, pos2 goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var posOnlyIDs, posVelIDs []uid.UID64
	var query *goke.Query
	var posQuery, velQuery *goke.Query
	var remover *goke.Remover
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		posOnly := si.NewFactory(&pos)
		posOnly.Create(2)
		for posOnly.Next() {
			posOnlyIDs = append(posOnlyIDs, posOnly.IDs...)
		}

		posVel := si.NewFactory(&pos2, &vel)
		posVel.Create(2)
		for posVel.Next() {
			posVelIDs = append(posVelIDs, posVel.IDs...)
		}

		// Only matches the Position-only archetype (excludes Velocity).
		query = si.NewQueryBuilder(&pos).Exclude(goke.Exclude[Velocity]()).Build()

		posQuery = si.NewQueryBuilder().Include(goke.Include[Position]()).Build()
		velQuery = si.NewQueryBuilder().Include(goke.Include[Velocity]()).Build()
		remover = si.Remover()
	}})

	sys := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			cursor := query.Cursor()
			if len(cursor.IDs) == 0 {
				continue
			}
			buf := query.BeginMigrate(cb)
			for _, id := range cursor.IDs {
				buf.Add(id)
			}
			buf.Commit(remover)
		}
	}})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range posOnlyIDs {
		if hasComp(posQuery, id) {
			t.Errorf("entity %v: expected already removed via cb.Commit(remover), but still matches", id)
		}
	}
	for _, id := range posVelIDs {
		if !hasComp(posQuery, id) {
			t.Errorf("entity %v: expected Position to remain (not matched by query)", id)
		}
		if !hasComp(velQuery, id) {
			t.Errorf("entity %v: expected Velocity to remain (not matched by query)", id)
		}
	}
}
