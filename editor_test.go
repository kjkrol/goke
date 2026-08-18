package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// TestEditorBuilder_AddComponent exercises the public Editor API end to
// end: query.NewEditorBuilder(comps...).Build(), applied via
// Query.BeginMigrate inside a system's Update, applied at Sync.
func TestEditorBuilder_AddComponent(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Velocity]()

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var ids []uid.UID64
	var query *goke.Query
	var addVel *goke.Editor
	var velQuery *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(2)
		for factory.Next() {
			ids = append(ids, factory.IDs...)
		}

		query = si.NewQueryBuilder(&pos).Exclude(goke.Exclude[Velocity]()).Build()
		addVel = query.NewEditorBuilder(&vel).Build()
		velQuery = si.NewQueryBuilder().Include(goke.Include[Velocity]()).Build()
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
			buf.Commit(addVel)
		}
	}})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if !hasComp(velQuery, id) {
			t.Errorf("entity %v: expected Velocity added via Editor", id)
		}
	}
}

// TestEditorBuilder_RemoveComponent exercises the Remove side: an Editor
// built via query.NewEditorBuilder().Remove(...).Build() removes a component
// from every matched entity.
func TestEditorBuilder_RemoveComponent(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Velocity]()

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var ids []uid.UID64
	var query *goke.Query
	var removeVel *goke.Editor
	var velQuery, posQuery *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos, &vel)
		factory.Create(2)
		for factory.Next() {
			ids = append(ids, factory.IDs...)
		}

		query = si.NewQueryBuilder(&pos, &vel).Build()
		removeVel = query.NewEditorBuilder().Remove(goke.Remove[Velocity]()).Build()
		velQuery = si.NewQueryBuilder().Include(goke.Include[Velocity]()).Build()
		posQuery = si.NewQueryBuilder().Include(goke.Include[Position]()).Build()
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
			buf.Commit(removeVel)
		}
	}})
	ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
		s.Run(sys, d)
		s.Sync()
	})

	ecs.Tick(time.Millisecond)

	for _, id := range ids {
		if hasComp(velQuery, id) {
			t.Errorf("entity %v: expected Velocity removed via Editor", id)
		}
		if !hasComp(posQuery, id) {
			t.Errorf("entity %v: expected Position to remain", id)
		}
	}
}
