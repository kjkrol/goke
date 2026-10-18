package goke_test

import (
	"slices"
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
)

type (
	farOff struct{}
	asleep struct{}
)

func TestQuery_PacedGoesOverEachBlockAsItsRuleSays(t *testing.T) {
	ecs := goke.New()
	var step uint64
	var pos goke.Comp[Position]
	var paced, plain *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		for _, f := range []*goke.Factory{
			si.NewFactory(new(goke.Comp[Position])),
			si.NewFactory(new(goke.Comp[Position]), new(goke.Comp[farOff])),
			si.NewFactory(new(goke.Comp[Position]), new(goke.Comp[asleep])),
		} {
			f.Create(1)
			f.Next()
		}
		paced = si.NewQueryBuilder(&pos).
			With(goke.PacedBy(func() uint64 { return step },
				goke.Every[farOff](3),
				goke.Never[asleep]())).
			Build()
		plain = si.NewQueryBuilder(&pos).Build()
	}})
	const d = time.Millisecond
	walked := func(q *goke.Query) (n int, dts []time.Duration) {
		for q.Paced(d); q.Next(); {
			n += len(q.Cursor().IDs)
			dts = append(dts, q.Dt())
		}
		return n, dts
	}
	var visits []int
	for step = 0; step < 6; step++ {
		n, dts := walked(paced)
		visits = append(visits, n)
		for _, dt := range dts {
			if dt != d && dt != 3*d {
				t.Errorf("step %d: a block given %v, want 1ms or 3ms", step, dt)
			}
		}
	}
	if want := []int{1, 1, 2, 1, 1, 2}; !slices.Equal(visits, want) {
		t.Errorf("entities walked by step %v, want %v", visits, want)
	}
	if n, dts := walked(plain); n != 3 || len(dts) != 3 || dts[0] != d {
		t.Errorf("an unpaced query walked %d entities over %v, want all three over 1ms", n, dts)
	}
	due := 0
	for paced.All(); paced.Next(); {
		if paced.Due() {
			due++
		}
	}
	if due != 1 {
		t.Errorf("at step 6, %d blocks due walking all, want the near one", due)
	}
}
