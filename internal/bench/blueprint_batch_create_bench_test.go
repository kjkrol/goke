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
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 2 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint2[Pos, Vel](ecs)

		fn := func() {
			index := 0
			for page := range blueprint.Create(batchSize) {
				for j, _ := range page.Entity {
					page.Comp1[j].X = 1
					page.Comp2[j].X = 2
					index++
				}
			}
		}

		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 3 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 4 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint4[Pos, Vel, Acc, T04](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
						page.Comp4[j].V = 4
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 5 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint5[Pos, Vel, Acc, T04, T05](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
						page.Comp4[j].V = 4
						page.Comp5[j].V = 5
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 6 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint6[Pos, Vel, Acc, T04, T05, T06](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
						page.Comp4[j].V = 4
						page.Comp5[j].V = 5
						page.Comp6[j].V = 6
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 7 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
						page.Comp4[j].V = 4
						page.Comp5[j].V = 5
						page.Comp6[j].V = 6
						page.Comp7[j].V = 7
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 8 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
						page.Comp4[j].V = 4
						page.Comp5[j].V = 5
						page.Comp6[j].V = 6
						page.Comp7[j].V = 7
						page.Comp8[j].V = 8
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 9 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
						page.Comp4[j].V = 4
						page.Comp5[j].V = 5
						page.Comp6[j].V = 6
						page.Comp7[j].V = 7
						page.Comp8[j].V = 8
						page.Comp9[j].V = 9
					}
				}
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 10 comp", batchSize), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		measurePerEntity(b, batchSize, func() {
			for i := 0; i < b.N; i++ {
				for page := range blueprint.Create(batchSize) {
					for j, _ := range page.Entity {
						page.Comp1[j].X = 1
						page.Comp2[j].X = 2
						page.Comp3[j].X = 3
						page.Comp4[j].V = 4
						page.Comp5[j].V = 5
						page.Comp6[j].V = 6
						page.Comp7[j].V = 7
						page.Comp8[j].V = 8
						page.Comp9[j].V = 9
						page.Comp10[j].V = 10
					}
				}
			}
		})
	})
}
