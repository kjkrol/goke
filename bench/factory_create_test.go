package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
)

func Benchmark_Factory_Create(b *testing.B) {
	ecs := setupECS()

	b.Run(fmt.Sprintf("pop=%d/1_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		factory := ecs.CreateFactory(goke.Add(&c1))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/2_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/3_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/4_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3), goke.Add(&c4))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/5_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3), goke.Add(&c4),
			goke.Add(&c5))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/6_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3), goke.Add(&c4),
			goke.Add(&c5), goke.Add(&c6))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/7_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3), goke.Add(&c4),
			goke.Add(&c5), goke.Add(&c6),
			goke.Add(&c7))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/8_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3), goke.Add(&c4),
			goke.Add(&c5), goke.Add(&c6),
			goke.Add(&c7), goke.Add(&c8))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/9_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		var c9 goke.Col[T09]
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3), goke.Add(&c4),
			goke.Add(&c5), goke.Add(&c6),
			goke.Add(&c7), goke.Add(&c8),
			goke.Add(&c9))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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

	b.Run(fmt.Sprintf("pop=%d/10_comp", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
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
		factory := ecs.CreateFactory(goke.Add(&c1), goke.Add(&c2),
			goke.Add(&c3), goke.Add(&c4),
			goke.Add(&c5), goke.Add(&c6),
			goke.Add(&c7), goke.Add(&c8),
			goke.Add(&c9), goke.Add(&c10))
		fc := &factory.Cursor
		fn := func() {
			factory.Create(entitiesNumber)
			for factory.Next() {
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
