package bench_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

func Benchmark_Remove(b *testing.B) {
	count := 100000
	b.Run(fmt.Sprintf("pop=%d", count), func(b *testing.B) {
		ecs := goke.New(
			goke.WithEntityCap(count),
			goke.WithEntityFreeCap(count),
		)
		_ = ecs.RegComp[Pos]()

		var factory *goke.Factory
		ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
			factory = si.NewFactory(new(goke.Comp[Pos]))
		}})
		entities := make([]uid.UID64, count)

		refill := func() {
			offset := 0
			factory.Create(b.N)
			for factory.Next() {
				n := copy(entities[offset:], factory.IDs)
				offset += n
			}
		}
		refill()

		var toRemove uid.UID64
		remover := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
			cb.RemoveOne(toRemove)
		}})
		ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
			ctx.Run(remover, d)
			_ = ctx.Sync()
		})

		measurePerEntity(b, 1, func() {
			for i := 0; b.Loop(); i++ {
				idx := i % count

				if i >= count && i%count == 0 {
					b.StopTimer()
					refill()
					b.StartTimer()
				}

				toRemove = entities[idx]
				ecs.Tick(0)
			}
		})
	})
}
