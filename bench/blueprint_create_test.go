package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
)

func Benchmark_Blueprint_Create(b *testing.B) {
	ecs := setupECS()

	b.Run(fmt.Sprintf("Batch(%d) 1 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint1[Pos](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 2 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint2[Pos, Vel](ecs)

		fn := func() {
			index := 0
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					index++
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 3 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 4 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint4[Pos, Vel, Acc, T04](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
					chunk.Comp4[j].V = 4
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 5 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint5[Pos, Vel, Acc, T04, T05](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
					chunk.Comp4[j].V = 4
					chunk.Comp5[j].V = 5
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 6 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint6[Pos, Vel, Acc, T04, T05, T06](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
					chunk.Comp4[j].V = 4
					chunk.Comp5[j].V = 5
					chunk.Comp6[j].V = 6
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 7 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
					chunk.Comp4[j].V = 4
					chunk.Comp5[j].V = 5
					chunk.Comp6[j].V = 6
					chunk.Comp7[j].V = 7
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 8 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
					chunk.Comp4[j].V = 4
					chunk.Comp5[j].V = 5
					chunk.Comp6[j].V = 6
					chunk.Comp7[j].V = 7
					chunk.Comp8[j].V = 8
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 9 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
					chunk.Comp4[j].V = 4
					chunk.Comp5[j].V = 5
					chunk.Comp6[j].V = 6
					chunk.Comp7[j].V = 7
					chunk.Comp8[j].V = 8
					chunk.Comp9[j].V = 9
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("Batch(%d) 10 comp", entitiesNumber), func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		fn := func() {
			for chunk := range blueprint.Create(entitiesNumber) {
				for j, _ := range chunk.Entity {
					chunk.Comp1[j].X = 1
					chunk.Comp2[j].X = 2
					chunk.Comp3[j].X = 3
					chunk.Comp4[j].V = 4
					chunk.Comp5[j].V = 5
					chunk.Comp6[j].V = 6
					chunk.Comp7[j].V = 7
					chunk.Comp8[j].V = 8
					chunk.Comp9[j].V = 9
					chunk.Comp10[j].V = 10
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})
}
