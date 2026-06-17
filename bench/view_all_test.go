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
//  2. Hoist the component slice once per chunk: `pos := col.Slice(&v.Cursor)`
//     outside the inner loop, then index `pos[i]`. Range cursor.EntSlice so the
//     compiler sees the same field driving both len(pos) and the loop bound; this
//     lets it prove `i < len(pos)` and eliminate bounds checks for any number of
//     tracked columns. Re-deriving the slice inside the loop, or indexing through
//     a struct field per use, loses that elimination and inflates the cost (~2.5x
//     at 3 comps).
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

	view0 := goke.CreateView(ecs)
	view1 := goke.CreateView(ecs, goke.Track(&pos))
	view2 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel))
	view3 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Include[T04]())
	view4 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04))
	view5 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05))
	view6 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06))
	view7 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07))
	view8 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08))
	view9 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09))
	view10 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09), goke.Track(&t10))

	var GlobalCount int
	b.Run("0 comp", func(b *testing.B) {
		fn := func() {
			count := 0
			view0.All()
			for view0.Next() {
				for _, entityID := range view0.Cursor.EntSlice {
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
		cursor := &view1.Cursor
		fn := func() {
			count := 0
			view1.All()
			for view1.Next() {
				posSlice := pos.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view2.Cursor
		fn := func() {
			view2.All()
			for view2.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view3.Cursor
		fn := func() {
			view3.All()
			for view3.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view4.Cursor
		fn := func() {
			view4.All()
			for view4.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view5.Cursor
		fn := func() {
			view5.All()
			for view5.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view6.Cursor
		fn := func() {
			view6.All()
			for view6.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view7.Cursor
		fn := func() {
			view7.All()
			for view7.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				t07Slice := t07.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view8.Cursor
		fn := func() {
			view8.All()
			for view8.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				t07Slice := t07.Slice(cursor)
				t08Slice := t08.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view9.Cursor
		fn := func() {
			view9.All()
			for view9.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				t07Slice := t07.Slice(cursor)
				t08Slice := t08.Slice(cursor)
				t09Slice := t09.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
		cursor := &view10.Cursor
		fn := func() {
			view10.All()
			for view10.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				t07Slice := t07.Slice(cursor)
				t08Slice := t08.Slice(cursor)
				t09Slice := t09.Slice(cursor)
				t10Slice := t10.Slice(cursor)
				for i, entityID := range cursor.EntSlice {
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
