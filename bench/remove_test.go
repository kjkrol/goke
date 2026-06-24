package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

func Benchmark_Remove(b *testing.B) {
	count := 100000
	b.Run(fmt.Sprintf("pop=%d", count), func(b *testing.B) {
		ecs := goke.New(
			goke.WithEntityCap(count),
			goke.WithEntityFreeCap(count),
		)
		_ = goke.RegComp[Pos](ecs)

		factory := ecs.NewFactory(new(goke.Comp[Pos]))
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

		measurePerEntity(b, 1, func() {
			for i := 0; b.Loop(); i++ {
				idx := i % count

				if i >= count && i%count == 0 {
					b.StopTimer()
					refill()
					b.StartTimer()
				}

				ecs.RemoveEnt(entities[idx])
			}
		})
	})
}
