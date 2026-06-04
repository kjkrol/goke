package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
)

func Benchmark_View_Filter100(b *testing.B) {

	ecs := setupECS()
	entities := populate(ecs, entitiesNumber)
	subset := entities[:100]

	b.Run("0 comp", func(b *testing.B) {
		view0 := goke.NewView0(ecs)
		fn := func() {
			for entity := range view0.Filter(subset) {
				_ = entity
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("1 comp", func(b *testing.B) {
		view3 := goke.NewView1[Pos](ecs)
		fn := func() {
			for item := range view3.Filter(subset) {
				pos := item.Comp1
				pos.X += pos.Y
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("2 comp", func(b *testing.B) {
		view3 := goke.NewView2[Pos, Vel](ecs)
		fn := func() {
			for item := range view3.Filter(subset) {
				pos, vel := item.Comp1, item.Comp2
				pos.X += vel.X
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("3 comp", func(b *testing.B) {
		view3 := goke.NewView3[Pos, Vel, Acc](ecs)
		fn := func() {
			for item := range view3.Filter(subset) {
				pos, vel, acc := item.Comp1, item.Comp2, item.Comp3
				acc.X += vel.X
				pos.X += vel.X
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})
}
