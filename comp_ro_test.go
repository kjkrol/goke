package goke_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type readMark struct{}

func spawnPositions(ecs *goke.ECS, n int, withVel bool) []uid.UID64 {
	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var mark goke.Comp[readMark]
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		comps := []goke.Addable{&pos}
		if withVel {
			comps = append(comps, &vel, &mark)
		}
		f := si.NewFactory(comps...)
		f.Create(n)
		for f.Next() {
			ps := pos.Slice(&f.Cursor)
			for i, id := range f.IDs {
				ps[i] = Position{X: float32(id), Y: 1}
				if withVel {
					vel.Slice(&f.Cursor)[i] = Velocity{VX: float32(id)}
				}
			}
			ids = append(ids, f.IDs...)
		}
	}})
	return ids
}

func TestCompRO_ReadsWhatCompHolds(t *testing.T) {
	ecs := goke.New()
	ids := spawnPositions(ecs, 300, false)

	var pos goke.CompRO[Position]
	var q *goke.Query
	ecs.RegSys(goke.SystemFn{OnInit: func(si *goke.SysInit) { q = si.NewQueryBuilder(&pos).Build() }})

	seen := 0
	cur := q.Cursor()
	for q.All(); q.Next(); {
		ps := pos.Slice(cur)
		if ps.Len() != len(cur.IDs) {
			t.Fatalf("expected Len %d, got %d", len(cur.IDs), ps.Len())
		}
		for i, id := range cur.IDs {
			if got := ps.At(i); got != (Position{X: float32(id), Y: 1}) {
				t.Fatalf("entity %d: expected X %d, got %+v", id, id, got)
			}
			seen++
		}
	}
	if seen != len(ids) {
		t.Errorf("expected %d entities, saw %d", len(ids), seen)
	}
	if !q.Seek(ids[7]) || pos.Get(cur) != (Position{X: float32(ids[7]), Y: 1}) {
		t.Errorf("expected Get to read entity %d's Position, got %+v", ids[7], pos.Get(cur))
	}
	if pos.Index() != 0 {
		t.Errorf("expected track position 0, got %d", pos.Index())
	}
	if toks := goke.LoadComps(&pos); len(toks) != 1 {
		t.Errorf("expected LoadComps to take a CompRO, got %d tokens", len(toks))
	}
}

func TestOptCompRO_PresentAndAbsent(t *testing.T) {
	ecs := goke.New()
	plain := spawnPositions(ecs, 3, false)[0]

	var vel goke.Comp[Velocity]
	var with uid.UID64
	var pos goke.Comp[Position]
	var mark goke.Comp[readMark]
	var optVel goke.OptCompRO[Velocity]
	var optMark goke.OptCompRO[readMark]
	var q *goke.Query
	ecs.RegSys(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&pos, &vel, &mark)
		f.Create(1)
		f.Next()
		with = f.IDs[0]
		vel.Slice(&f.Cursor)[0] = Velocity{VX: 5}
		q = si.NewQueryBuilder(&goke.CompRO[Position]{}).Optional(&optVel, &optMark).Build()
	}})

	cur := q.Cursor()
	for q.All(); q.Next(); {
		want := 0
		if optVel.Present(cur) {
			want = len(cur.IDs)
		}
		if optMark.Present(cur) != optVel.Present(cur) {
			t.Fatalf("expected Velocity and the tag together")
		}
		if vs, ms := optVel.Slice(cur), optMark.Slice(cur); vs.Len() != want || ms.Len() != want {
			t.Errorf("expected Len %d, got %d and %d", want, vs.Len(), ms.Len())
		}
	}

	if optVel.Index() != 0 || optMark.Index() != 1 {
		t.Errorf("expected optional positions 0 and 1, got %d and %d", optVel.Index(), optMark.Index())
	}
	q.Seek(with)
	if v, ok := optVel.Get(cur); !ok || v.VX != 5 {
		t.Errorf("expected Velocity{5} present, got %+v %v", v, ok)
	}
	if _, ok := optMark.Get(cur); !ok {
		t.Errorf("expected the tag present")
	}
	q.Seek(plain)
	if v, ok := optVel.Get(cur); ok || v != (Velocity{}) {
		t.Errorf("expected no Velocity, got %+v %v", v, ok)
	}
}

func TestRunParallel_ChecksQueryAccess(t *testing.T) {
	posType := reflect.TypeFor[Position]()
	velType := reflect.TypeFor[Velocity]()
	sys := func(comps ...goke.Trackable) goke.System {
		return goke.SystemFn{OnInit: func(si *goke.SysInit) { si.NewQueryBuilder(comps...).Build() }}
	}
	optSys := func(comps ...goke.OptTrackable) goke.System {
		return goke.SystemFn{OnInit: func(si *goke.SysInit) { si.NewQueryBuilder().Optional(comps...).Build() }}
	}

	cases := []struct {
		name  string
		a, b  goke.System
		panic string
	}{
		{"readers", sys(&goke.CompRO[Position]{}), sys(goke.NewColumnRO(posType)), ""},
		{"disjoint writers", sys(&goke.Comp[Position]{}), sys(goke.NewColumn(velType)), ""},
		{"writer and reader", sys(&goke.Comp[Position]{}), sys(&goke.CompRO[Position]{}), "conflict on goke_test.Position"},
		{"optional reader", optSys(&goke.OptCompRO[Velocity]{}), optSys(goke.NewOptColumnRO(velType)), ""},
		{"optional writer", optSys(&goke.OptComp[Velocity]{}), sys(&goke.CompRO[Velocity]{}), "conflict on goke_test.Velocity"},
		{"filters only", goke.SystemFn{OnInit: func(si *goke.SysInit) {
			si.NewQueryBuilder().Include(goke.Include[Position]()).Build()
		}}, sys(&goke.Comp[Position]{}), ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ecs := goke.New()
			spawnPositions(ecs, 1, true)
			a, b := ecs.RegSys(c.a), ecs.RegSys(c.b)
			ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) { ctx.RunParallel(d, a, b) })

			defer func() {
				msg := ""
				if p := recover(); p != nil {
					msg = fmt.Sprint(p)
				}
				if c.panic == "" && msg != "" || !strings.Contains(msg, c.panic) {
					t.Fatalf("expected panic %q, got %q", c.panic, msg)
				}
			}()
			ecs.Tick(time.Millisecond)
		})
	}
}
