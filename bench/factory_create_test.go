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
		var c1 goke.Comp[Pos]
		factory := ecs.NewFactory(&c1)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		factory := ecs.NewFactory(&c1, &c2)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		factory := ecs.NewFactory(&c1, &c2, &c3)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		factory := ecs.NewFactory(&c1, &c2, &c3, &c4)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		factory := ecs.NewFactory(&c1, &c2, &c3, &c4, &c5)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		factory := ecs.NewFactory(&c1, &c2, &c3, &c4, &c5, &c6)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		factory := ecs.NewFactory(&c1, &c2, &c3, &c4, &c5, &c6, &c7)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		var c8 goke.Comp[T08]
		factory := ecs.NewFactory(&c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		var c8 goke.Comp[T08]
		var c9 goke.Comp[T09]
		factory := ecs.NewFactory(&c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9)
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		var c8 goke.Comp[T08]
		var c9 goke.Comp[T09]
		var c10 goke.Comp[T10]
		factory := ecs.NewFactory(&c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10)
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
