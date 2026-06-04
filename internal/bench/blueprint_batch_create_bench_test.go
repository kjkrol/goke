package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
)

func Benchmark_Blueprint_BatchCreate(b *testing.B) {
	ecs := setupECS()

	const batchSize = 1024

	b.Run(fmt.Sprintf("Batch(%d) 1 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint1[Pos](ecs)
		buf := make([]goke.Item1[Pos], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 2 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint2[Pos, Vel](ecs)
		buf := make([]goke.Item2[Pos, Vel], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
					item.Comp2.X = 2
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 3 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)
		buf := make([]goke.Item3[Pos, Vel, Acc], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
					item.Comp2.X = 2
					item.Comp3.X = 3
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 4 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint4[Pos, Vel, Acc, T04](ecs)
		buf := make([]goke.Item4[Pos, Vel, Acc, T04], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
					item.Comp2.X = 2
					item.Comp3.X = 3
					item.Comp4.V = 4
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 5 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint5[Pos, Vel, Acc, T04, T05](ecs)
		buf := make([]goke.Item5[Pos, Vel, Acc, T04, T05], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
					item.Comp2.X = 2
					item.Comp3.X = 3
					item.Comp4.V = 4
					item.Comp5.V = 5
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 6 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint6[Pos, Vel, Acc, T04, T05, T06](ecs)
		buf := make([]goke.Item6[Pos, Vel, Acc, T04, T05, T06], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
					item.Comp2.X = 2
					item.Comp3.X = 3
					item.Comp4.V = 4
					item.Comp5.V = 5
					item.Comp6.V = 6
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 7 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		buf := make([]goke.Item7[Pos, Vel, Acc, T04, T05, T06, T07], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
					item.Comp2.X = 2
					item.Comp3.X = 3
					item.Comp4.V = 4
					item.Comp5.V = 5
					item.Comp6.V = 6
					item.Comp7.V = 7
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 8 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		buf := make([]goke.Item8[Pos, Vel, Acc, T04, T05, T06, T07, T08], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
					item.Comp1.X = 1
					item.Comp2.X = 2
					item.Comp3.X = 3
					item.Comp4.V = 4
					item.Comp5.V = 5
					item.Comp6.V = 6
					item.Comp7.V = 7
					item.Comp8.V = 8
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 9 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		buf := make([]goke.Item9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
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
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 10 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		buf := make([]goke.Item10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10], batchSize)
		blueprint.BatchCreate(batchSize, buf)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for _, item := range blueprint.BatchCreate(batchSize, buf) {
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
			}
		})
	})
}
