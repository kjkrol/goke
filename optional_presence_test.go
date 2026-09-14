package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// Optional must report presence on a query with no required components, for a
// component added at runtime — i.e. in an archetype built after the query.
func TestQuery_OptionalReportsPresenceOnQueryWithNoRequiredComponents(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()
	velID := ecs.RegComp[Velocity]()

	var pos goke.Comp[Position]
	var velOpt goke.OptComp[Velocity]
	var velReq goke.Comp[Velocity]
	var optQ, reqQ *goke.Query
	var id uid.UID64

	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&pos)
		f.Create(1)
		f.Next()
		id = f.IDs[0]

		optQ = si.NewQueryBuilder().Optional(&velOpt).Build()
		reqQ = si.NewQueryBuilder(&velReq).Build()
	}})

	adder := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
		cb.AddOne(id, velID, Velocity{VX: 3})
	}})
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(adder, d)
		_ = ctx.Sync()
	})
	ecs.Tick(0)

	reqSeen := 0
	reqQ.All()
	for reqQ.Next() {
		reqSeen += len(reqQ.Cursor().IDs)
	}
	if reqSeen != 1 {
		t.Fatalf("required-component query saw %d entities with Velocity, want 1 — the add did not land", reqSeen)
	}

	optPresent, optSliceLen := 0, 0
	optQ.All()
	for optQ.Next() {
		cur := optQ.Cursor()
		if !velOpt.Present(cur) {
			continue
		}
		optPresent += len(cur.IDs)
		optSliceLen += len(velOpt.Slice(cur))
	}
	if optPresent != 1 {
		t.Errorf("Optional query reported Present for %d entities, want 1 (the required-component query found it)", optPresent)
	}
	if optSliceLen != 1 {
		t.Errorf("Optional query's Slice covered %d elements, want 1", optSliceLen)
	}
}

// The same shape, but the optional component exists from the start.
func TestQuery_OptionalReportsPresenceForPreexistingArchetype(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Velocity]()

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var velOpt goke.OptComp[Velocity]
	var optQ *goke.Query

	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&pos, &vel)
		f.Create(1)
		f.Next()
		optQ = si.NewQueryBuilder().Optional(&velOpt).Build()
	}})

	seen := 0
	optQ.All()
	for optQ.Next() {
		if cur := optQ.Cursor(); velOpt.Present(cur) {
			seen += len(cur.IDs)
		}
	}
	if seen != 1 {
		t.Errorf("Optional query reported Present for %d entities, want 1", seen)
	}
}

// Two Optional columns on one query must not read each other's slots.
func TestQuery_TwoOptionalsReportIndependently(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Velocity]()
	_ = ecs.RegComp[Health]()

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var velOpt goke.OptComp[Velocity]
	var healthOpt goke.OptComp[Health]
	var q *goke.Query

	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&pos, &vel)
		f.Create(1)
		f.Next()
		vel.Slice(&f.Cursor)[0] = Velocity{VX: 5}

		q = si.NewQueryBuilder().Optional(&velOpt).Optional(&healthOpt).Build()
	}})

	sawVel, sawHealth := 0, 0
	var gotVX float32
	q.All()
	for q.Next() {
		cur := q.Cursor()
		if velOpt.Present(cur) {
			sawVel += len(cur.IDs)
			gotVX = velOpt.Slice(cur)[0].VX
		}
		if healthOpt.Present(cur) {
			sawHealth += len(cur.IDs)
		}
	}
	if sawVel != 1 {
		t.Errorf("Velocity reported present for %d entities, want 1", sawVel)
	}
	if gotVX != 5 {
		t.Errorf("Velocity.VX = %v, want 5 — the optional slice addressed the wrong column", gotVX)
	}
	if sawHealth != 0 {
		t.Errorf("Health reported present for %d entities, want 0 — nothing has it", sawHealth)
	}
}
