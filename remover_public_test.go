package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

func TestRemover_RemovesMatchingEntities(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Comp[Position]
	factory := ecs.NewFactory(&pos)
	factory.Create(3)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	query := ecs.NewQueryBuilder(&pos).Build()

	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
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
			buf.CommitRemove()
		}
	})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if ecs.RemoveEnt(id) {
			t.Errorf("entity %v: expected already removed via cb.Remove, but RemoveEnt succeeded", id)
		}
	}
}

func TestRemover_LeavesNonMatchingEntitiesIntact(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	posOnly := ecs.NewFactory(&pos)
	posOnly.Create(2)
	var posOnlyIDs []uid.UID64
	for posOnly.Next() {
		posOnlyIDs = append(posOnlyIDs, posOnly.IDs...)
	}

	var pos2 goke.Comp[Position]
	var vel goke.Comp[Velocity]
	posVel := ecs.NewFactory(&pos2, &vel)
	posVel.Create(2)
	var posVelIDs []uid.UID64
	for posVel.Next() {
		posVelIDs = append(posVelIDs, posVel.IDs...)
	}

	// Only matches the Position-only archetype (excludes Velocity).
	query := ecs.NewQueryBuilder(&pos).Exclude(goke.Exclude[Velocity]()).Build()

	sys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
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
			buf.CommitRemove()
		}
	})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range posOnlyIDs {
		if ecs.RemoveEnt(id) {
			t.Errorf("entity %v: expected already removed via cb.Remove, but RemoveEnt succeeded", id)
		}
	}
	for _, id := range posVelIDs {
		if !hasComp[Position](ecs, id) {
			t.Errorf("entity %v: expected Position to remain (not matched by query)", id)
		}
		if !hasComp[Velocity](ecs, id) {
			t.Errorf("entity %v: expected Velocity to remain (not matched by query)", id)
		}
	}
}
