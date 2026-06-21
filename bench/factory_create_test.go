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
		var c1 goke.Col[Pos]
		blueprint := goke.CreateEntFactory(ecs, goke.Track(&c1))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3), goke.Track(&c4))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				comp4 := c4.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
					comp4[j].V = 4
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3), goke.Track(&c4),
			goke.Track(&c5))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				comp4 := c4.Slice(fc)
				comp5 := c5.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
					comp4[j].V = 4
					comp5[j].V = 5
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3), goke.Track(&c4),
			goke.Track(&c5), goke.Track(&c6))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				comp4 := c4.Slice(fc)
				comp5 := c5.Slice(fc)
				comp6 := c6.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
					comp4[j].V = 4
					comp5[j].V = 5
					comp6[j].V = 6
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3), goke.Track(&c4),
			goke.Track(&c5), goke.Track(&c6),
			goke.Track(&c7))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				comp4 := c4.Slice(fc)
				comp5 := c5.Slice(fc)
				comp6 := c6.Slice(fc)
				comp7 := c7.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
					comp4[j].V = 4
					comp5[j].V = 5
					comp6[j].V = 6
					comp7[j].V = 7
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3), goke.Track(&c4),
			goke.Track(&c5), goke.Track(&c6),
			goke.Track(&c7), goke.Track(&c8))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				comp4 := c4.Slice(fc)
				comp5 := c5.Slice(fc)
				comp6 := c6.Slice(fc)
				comp7 := c7.Slice(fc)
				comp8 := c8.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
					comp4[j].V = 4
					comp5[j].V = 5
					comp6[j].V = 6
					comp7[j].V = 7
					comp8[j].V = 8
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		var c9 goke.Col[T09]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3), goke.Track(&c4),
			goke.Track(&c5), goke.Track(&c6),
			goke.Track(&c7), goke.Track(&c8),
			goke.Track(&c9))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				comp4 := c4.Slice(fc)
				comp5 := c5.Slice(fc)
				comp6 := c6.Slice(fc)
				comp7 := c7.Slice(fc)
				comp8 := c8.Slice(fc)
				comp9 := c9.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
					comp4[j].V = 4
					comp5[j].V = 5
					comp6[j].V = 6
					comp7[j].V = 7
					comp8[j].V = 8
					comp9[j].V = 9
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
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		var c9 goke.Col[T09]
		var c10 goke.Col[T10]
		blueprint := goke.CreateEntFactory(ecs,
			goke.Track(&c1), goke.Track(&c2),
			goke.Track(&c3), goke.Track(&c4),
			goke.Track(&c5), goke.Track(&c6),
			goke.Track(&c7), goke.Track(&c8),
			goke.Track(&c9), goke.Track(&c10))
		fc := &blueprint.Cursor
		fn := func() {
			blueprint.Create(entitiesNumber)
			for blueprint.Next() {
				comp1 := c1.Slice(fc)
				comp2 := c2.Slice(fc)
				comp3 := c3.Slice(fc)
				comp4 := c4.Slice(fc)
				comp5 := c5.Slice(fc)
				comp6 := c6.Slice(fc)
				comp7 := c7.Slice(fc)
				comp8 := c8.Slice(fc)
				comp9 := c9.Slice(fc)
				comp10 := c10.Slice(fc)
				for j := range fc.IDs {
					comp1[j].X = 1
					comp2[j].X = 2
					comp3[j].X = 3
					comp4[j].V = 4
					comp5[j].V = 5
					comp6[j].V = 6
					comp7[j].V = 7
					comp8[j].V = 8
					comp9[j].V = 9
					comp10[j].V = 10
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
