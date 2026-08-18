package bench_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// Benchmark_Stability_Grow exercises interleaved entity creation and removal
// under unbounded growth (1 created per iteration, only every other one
// removed — net population grows for the whole run). It measures whether
// generation-based ID recycling and the address book stay cheap and stable
// under that churn, not throughput on a fixed-size population.
func Benchmark_Stability_Grow(b *testing.B) {
	ecs := goke.New(goke.WithEntityCap(1024))
	_ = ecs.RegComp[Pos]()
	var pos goke.Comp[Pos]
	var factory *goke.Factory
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory = si.NewFactory(&pos)
	}})
	fc := &factory.Cursor

	var toRemove uid.UID64
	var shouldRemove bool
	remover := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
		if shouldRemove {
			cb.RemoveOne(toRemove)
		}
	}})
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(remover, d)
		_ = ctx.Sync()
	})

	var e uid.UID64
	measurePerEntity(b, 1, func() {
		for i := 0; b.Loop(); i++ {
			factory.Create(1)
			factory.Next()
			e = factory.IDs[0]
			pos.Slice(fc)[0] = Pos{X: 1}

			shouldRemove = i%2 == 0
			toRemove = e
			ecs.Tick(0)
		}
	})
}
