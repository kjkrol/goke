package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
)

// Benchmark_Matcher_All measures full chunk iteration via Matcher.All()/Next(),
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
//     outside the inner loop, then index `pos[i]`. Range cursor.IDs so the
//     compiler sees the same field driving both len(pos) and the loop bound; this
//     lets it prove `i < len(pos)` and eliminate bounds checks for any number of
//     tracked columns. Re-deriving the slice inside the loop, or indexing through
//     a struct field per use, loses that elimination and inflates the cost (~2.5x
//     at 3 comps).
//
// Matchers are created once outside b.Run: with -count=N each b.Run callback is
// called N times, so creating a new Matcher inside would accumulate N matchers per
// sub-benchmark on the same ECS and eventually exceed MaxMatchers.
func Benchmark_Matcher_All(b *testing.B) {
	ecs := setupECS()
	populate(ecs, entitiesNumber)

	var pos goke.Comp[Pos]
	var vel goke.Comp[Vel]
	var acc goke.Comp[Acc]
	var t04 goke.Comp[T04]
	var t05 goke.Comp[T05]
	var t06 goke.Comp[T06]
	var t07 goke.Comp[T07]
	var t08 goke.Comp[T08]
	var t09 goke.Comp[T09]
	var t10 goke.Comp[T10]

	matcher0 := ecs.NewQueryBuilder().Build()
	matcher1 := ecs.NewQueryBuilder(&pos).Build()
	matcher2 := ecs.NewQueryBuilder(&pos, &vel).Build()
	matcher3 := ecs.NewQueryBuilder(&pos, &vel, &acc).Include(goke.Include[T04]()).Build()
	matcher4 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04).Build()
	matcher5 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05).Build()
	matcher6 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06).Build()
	matcher7 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07).Build()
	matcher8 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08).Build()
	matcher9 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09).Build()
	matcher10 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09, &t10).Build()

	var GlobalCount int
	b.Run(fmt.Sprintf("pop=%d/0_comp", entitiesNumber), func(b *testing.B) {
		fn := func() {
			count := 0
			matcher0.All()
			for matcher0.Next() {
				for _, entityID := range matcher0.Cursor.IDs {
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
			b.Fatalf("Matcher0 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
		}
	})

	b.Run(fmt.Sprintf("pop=%d/1_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher1.Cursor
		fn := func() {
			count := 0
			matcher1.All()
			for matcher1.Next() {
				posSlice := pos.Slice(cursor)
				for i, entityID := range cursor.IDs {
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
			b.Fatalf("Matcher1 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
		}
	})

	b.Run(fmt.Sprintf("pop=%d/2_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher2.Cursor
		fn := func() {
			matcher2.All()
			for matcher2.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/3_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher3.Cursor
		fn := func() {
			matcher3.All()
			for matcher3.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/4_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher4.Cursor
		fn := func() {
			matcher4.All()
			for matcher4.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/5_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher5.Cursor
		fn := func() {
			matcher5.All()
			for matcher5.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/6_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher6.Cursor
		fn := func() {
			matcher6.All()
			for matcher6.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/7_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher7.Cursor
		fn := func() {
			matcher7.All()
			for matcher7.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				t07Slice := t07.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/8_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher8.Cursor
		fn := func() {
			matcher8.All()
			for matcher8.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				t07Slice := t07.Slice(cursor)
				t08Slice := t08.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/9_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher9.Cursor
		fn := func() {
			matcher9.All()
			for matcher9.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				t04Slice := t04.Slice(cursor)
				t05Slice := t05.Slice(cursor)
				t06Slice := t06.Slice(cursor)
				t07Slice := t07.Slice(cursor)
				t08Slice := t08.Slice(cursor)
				t09Slice := t09.Slice(cursor)
				for i, entityID := range cursor.IDs {
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

	b.Run(fmt.Sprintf("pop=%d/10_comp", entitiesNumber), func(b *testing.B) {
		cursor := &matcher10.Cursor
		fn := func() {
			matcher10.All()
			for matcher10.Next() {
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
				for i, entityID := range cursor.IDs {
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
