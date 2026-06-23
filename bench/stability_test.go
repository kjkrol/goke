package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

// Benchmark_Stability_Grow exercises interleaved entity creation and removal
// under unbounded growth (1 created per iteration, only every other one
// removed — net population grows for the whole run). It measures whether
// generation-based ID recycling and the address book stay cheap and stable
// under that churn, not throughput on a fixed-size population.
func Benchmark_Stability_Grow(b *testing.B) {
	ecs := goke.New(goke.WithEntityCap(1024))
	_ = goke.RegComp[Pos](ecs)
	var pos goke.Comp[Pos]
	factory := ecs.NewFactory(&pos)
	fc := &factory.Cursor

	var e uid.UID64
	measurePerEntity(b, 1, func() {
		for i := 0; b.Loop(); i++ {
			factory.Create(1)
			factory.Next()
			e = factory.IDs[0]
			pos.Slice(fc)[0] = Pos{X: 1}

			if i%2 == 0 {
				ecs.RemoveEnt(e)
			}
		}
	})
}
