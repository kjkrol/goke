package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
)

func Benchmark_Blueprint_Create(b *testing.B) {
	ecs := setupECS()

	b.Run("1 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint1[Pos](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
			}
		})

	})
	b.Run("2 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint2[Pos, Vel](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
			}
		})
	})
	b.Run("3 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
			}
		})
	})
	b.Run("4 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint4[Pos, Vel, Acc, T04](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
				item.Comp4.V = 4
			}
		})
	})
	b.Run("5 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint5[Pos, Vel, Acc, T04, T05](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
				item.Comp4.V = 4
				item.Comp5.V = 5
			}
		})
	})
	b.Run("6 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint6[Pos, Vel, Acc, T04, T05, T06](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
				item.Comp4.V = 4
				item.Comp5.V = 5
				item.Comp6.V = 6
			}
		})
	})
	b.Run("7 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
				item.Comp4.V = 4
				item.Comp5.V = 5
				item.Comp6.V = 6
				item.Comp7.V = 7
			}
		})
	})
	b.Run("8 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
				item.Comp4.V = 4
				item.Comp5.V = 5
				item.Comp6.V = 6
				item.Comp7.V = 7
				item.Comp8.V = 8
			}
		})
	})
	b.Run("9 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
				item.Comp4.V = 4
				item.Comp5.V = 5
				item.Comp6.V = 6
				item.Comp7.V = 7
				item.Comp8.V = 8
				item.Comp9.V = 9
			}
		})
	})
	b.Run("10 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				item := blueprint.Create()
				_ = item.Entity
				item.Comp1.X = 1
				item.Comp2.X = 2
				item.Comp3.X = 3
				item.Comp4.V = 4
				item.Comp5.V = 5
				item.Comp6.V = 6
				item.Comp7.V = 7
				item.Comp8.V = 8
				item.Comp9.V = 9
				item.Comp10.V = 10
			}
		})
	})
}
