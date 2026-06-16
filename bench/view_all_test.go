package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
)

// Benchmark_View_All measures full chunk iteration via View.All()/Next(),
// reading 0..10 component columns per entity.
//
// Two rules keep these numbers meaningful — do not "simplify" them away:
//
//  1. Write in place: `pos[i].X += pos[i].Y`, never `p := pos[i]; p.X += p.Y`.
//     A copy-then-mutate touches only a local that is never read again, so the
//     compiler deletes the write (dead-store elimination) and the benchmark
//     measures iteration over nothing — artificially ~2x too fast at low
//     component counts. In-place indexing forces a real store to chunk memory.
//
//  2. Hoist the component slice once per chunk: `pos := col.Slice(v)`
//     outside the inner loop, then index `pos[i]`. Because len(pos) == len(v.EntSlice)
//     (both derive from the chunk length), ranging v.EntSlice lets the compiler
//     prove `i < len(pos)` and eliminate the bounds check on pos[i]. Re-deriving
//     the slice inside the loop, or indexing through a struct field per use,
//     loses that bounds-check elimination and inflates the cost (~2.5x at 3 comps).
//
// Views are created once outside b.Run: with -count=N each b.Run callback is
// called N times, so creating a NewView inside would accumulate N views per
// sub-benchmark on the same ECS and eventually exceed MaxViews.
func Benchmark_View_All(b *testing.B) {
	ecs := setupECS()
	populate(ecs, entitiesNumber)

	var pos goke.Col[Pos]
	var vel goke.Col[Vel]
	var acc goke.Col[Acc]
	var t04 goke.Col[T04]
	var t05 goke.Col[T05]
	var t06 goke.Col[T06]
	var t07 goke.Col[T07]
	var t08 goke.Col[T08]
	var t09 goke.Col[T09]
	var t10 goke.Col[T10]

	view0 := goke.NewView(ecs)
	view1 := goke.NewView(ecs, pos.Track())
	view2 := goke.NewView(ecs, pos.Track(), vel.Track())
	view3 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), goke.Include[T04]())
	view4 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track())
	view5 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track())
	view6 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track())
	view7 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track())
	view8 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track(), t08.Track())
	view9 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track(), t08.Track(), t09.Track())
	view10 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track(), t08.Track(), t09.Track(), t10.Track())

	var GlobalCount int
	b.Run("0 comp", func(b *testing.B) {
		fn := func() {
			count := 0
			view0.All()
			for view0.Next() {
				for _, entityID := range view0.EntSlice {
					_ = entityID
					count++
				}
			}
			GlobalCount = count
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})

		if GlobalCount != entitiesNumber {
			b.Fatalf("View0 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
		}
	})

	b.Run("1 comp", func(b *testing.B) {
		fn := func() {
			count := 0
			view1.All()
			for view1.Next() {
				posSlice := pos.Slice(view1)
				for i, entityID := range view1.EntSlice {
					_ = entityID
					posSlice[i].X += posSlice[i].Y
					count++
				}
			}
			GlobalCount = count
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})

		if GlobalCount != entitiesNumber {
			b.Fatalf("View1 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
		}
	})

	b.Run("2 comp", func(b *testing.B) {
		fn := func() {
			view2.All()
			for view2.Next() {
				posSlice := pos.Slice(view2)
				velSlice := vel.Slice(view2)
				for i, entityID := range view2.EntSlice {
					_ = entityID
					velSlice[i].X += velSlice[i].Y
					posSlice[i].X += velSlice[i].X
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("3 comp", func(b *testing.B) {
		fn := func() {
			view3.All()
			for view3.Next() {
				posSlice := pos.Slice(view3)
				velSlice := vel.Slice(view3)
				accSlice := acc.Slice(view3)
				for i, entityID := range view3.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("4 comp", func(b *testing.B) {
		fn := func() {
			view4.All()
			for view4.Next() {
				posSlice := pos.Slice(view4)
				velSlice := vel.Slice(view4)
				accSlice := acc.Slice(view4)
				t04Slice := t04.Slice(view4)
				for i, entityID := range view4.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
					t04Slice[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("5 comp", func(b *testing.B) {
		fn := func() {
			view5.All()
			for view5.Next() {
				posSlice := pos.Slice(view5)
				velSlice := vel.Slice(view5)
				accSlice := acc.Slice(view5)
				t04Slice := t04.Slice(view5)
				t05Slice := t05.Slice(view5)
				for i, entityID := range view5.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
					t04Slice[i].V = 1
					t05Slice[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("6 comp", func(b *testing.B) {
		fn := func() {
			view6.All()
			for view6.Next() {
				posSlice := pos.Slice(view6)
				velSlice := vel.Slice(view6)
				accSlice := acc.Slice(view6)
				t04Slice := t04.Slice(view6)
				t05Slice := t05.Slice(view6)
				t06Slice := t06.Slice(view6)
				for i, entityID := range view6.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
					t04Slice[i].V = 1
					t05Slice[i].V = 1
					t06Slice[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("7 comp", func(b *testing.B) {
		fn := func() {
			view7.All()
			for view7.Next() {
				posSlice := pos.Slice(view7)
				velSlice := vel.Slice(view7)
				accSlice := acc.Slice(view7)
				t04Slice := t04.Slice(view7)
				t05Slice := t05.Slice(view7)
				t06Slice := t06.Slice(view7)
				t07Slice := t07.Slice(view7)
				for i, entityID := range view7.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
					t04Slice[i].V = 1
					t05Slice[i].V = 1
					t06Slice[i].V = 1
					t07Slice[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("8 comp", func(b *testing.B) {
		fn := func() {
			view8.All()
			for view8.Next() {
				posSlice := pos.Slice(view8)
				velSlice := vel.Slice(view8)
				accSlice := acc.Slice(view8)
				t04Slice := t04.Slice(view8)
				t05Slice := t05.Slice(view8)
				t06Slice := t06.Slice(view8)
				t07Slice := t07.Slice(view8)
				t08Slice := t08.Slice(view8)
				for i, entityID := range view8.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
					t04Slice[i].V = 1
					t05Slice[i].V = 1
					t06Slice[i].V = 1
					t07Slice[i].V = 1
					t08Slice[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("9 comp", func(b *testing.B) {
		fn := func() {
			view9.All()
			for view9.Next() {
				posSlice := pos.Slice(view9)
				velSlice := vel.Slice(view9)
				accSlice := acc.Slice(view9)
				t04Slice := t04.Slice(view9)
				t05Slice := t05.Slice(view9)
				t06Slice := t06.Slice(view9)
				t07Slice := t07.Slice(view9)
				t08Slice := t08.Slice(view9)
				t09Slice := t09.Slice(view9)
				for i, entityID := range view9.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
					t04Slice[i].V = 1
					t05Slice[i].V = 1
					t06Slice[i].V = 1
					t07Slice[i].V = 1
					t08Slice[i].V = 1
					t09Slice[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("10 comp", func(b *testing.B) {
		fn := func() {
			view10.All()
			for view10.Next() {
				posSlice := pos.Slice(view10)
				velSlice := vel.Slice(view10)
				accSlice := acc.Slice(view10)
				t04Slice := t04.Slice(view10)
				t05Slice := t05.Slice(view10)
				t06Slice := t06.Slice(view10)
				t07Slice := t07.Slice(view10)
				t08Slice := t08.Slice(view10)
				t09Slice := t09.Slice(view10)
				t10Slice := t10.Slice(view10)
				for i, entityID := range view10.EntSlice {
					_ = entityID
					accSlice[i].X += 0.1
					velSlice[i].X += accSlice[i].X
					posSlice[i].X += velSlice[i].X
					t04Slice[i].V = 1
					t05Slice[i].V = 1
					t06Slice[i].V = 1
					t07Slice[i].V = 1
					t08Slice[i].V = 1
					t09Slice[i].V = 1
					t10Slice[i].V = 1
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
