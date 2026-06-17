package bench_test

import (
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

// Benchmark_View_Filter measures the performance of the View.Filter method
// across different query configurations (0 to 10 components).
// Each sub-benchmark runs the same N=100 subset under two access patterns:
//   - sorted   : entities in creation order (best case for the per-entityID
//     cached archetype descriptor)
//   - shuffled : the same handles but rand.Shuffle'd (forces frequent
//     archetype-descriptor reloads)
//
// Filter yields pointers to live component memory. All variants must
// report 0 B/op and 0 allocs/op.
//
// Views are created once outside b.Run: with -count=N each b.Run callback is
// called N times, so creating a NewView inside would accumulate N views per
// sub-benchmark on the same ECS and eventually exceed MaxViews.
func Benchmark_View_Filter(b *testing.B) {
	ecs := setupECS()
	entities := populate(ecs, entitiesNumber)

	sortedSubset := append([]uid.UID64(nil), entities[:filterSubsetSize]...)
	shuffledSubset := append([]uid.UID64(nil), entities[:filterSubsetSize]...)
	rng := rand.New(rand.NewPCG(42, 1337))
	rng.Shuffle(len(shuffledSubset), func(i, j int) {
		shuffledSubset[i], shuffledSubset[j] = shuffledSubset[j], shuffledSubset[i]
	})

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
	view3 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc))
	view4 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04))
	view5 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05))
	view6 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06))
	view7 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07))
	view8 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08))
	view9 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09))
	view10 := goke.CreateView(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09), goke.Track(&t10))

	// --- 0 comp ---
	b.Run("0_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view0.Filter(sortedSubset)
				for view0.Next() {
				}
			}
		})
	})
	b.Run("0_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view0.Filter(shuffledSubset)
				for view0.Next() {
				}
			}
		})
	})

	// --- 1 comp ---
	b.Run("1_comp/sorted", func(b *testing.B) {
		cursor := &view1.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view1.Filter(sortedSubset)
				for view1.Next() {
					pos.At(cursor).X += pos.At(cursor).Y
				}
			}
		})
	})
	b.Run("1_comp/shuffled", func(b *testing.B) {
		cursor := &view1.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view1.Filter(shuffledSubset)
				for view1.Next() {
					pos.At(cursor).X += pos.At(cursor).Y
				}
			}
		})
	})

	// --- 2 comp ---
	b.Run("2_comp/sorted", func(b *testing.B) {
		cursor := &view2.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view2.Filter(sortedSubset)
				for view2.Next() {
					pos.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})
	b.Run("2_comp/shuffled", func(b *testing.B) {
		cursor := &view2.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view2.Filter(shuffledSubset)
				for view2.Next() {
					pos.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})

	// --- 3 comp ---
	b.Run("3_comp/sorted", func(b *testing.B) {
		cursor := &view3.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view3.Filter(sortedSubset)
				for view3.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})
	b.Run("3_comp/shuffled", func(b *testing.B) {
		cursor := &view3.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view3.Filter(shuffledSubset)
				for view3.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})

	// --- 4 comp ---
	b.Run("4_comp/sorted", func(b *testing.B) {
		cursor := &view4.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view4.Filter(sortedSubset)
				for view4.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
				}
			}
		})
	})
	b.Run("4_comp/shuffled", func(b *testing.B) {
		cursor := &view4.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view4.Filter(shuffledSubset)
				for view4.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
				}
			}
		})
	})

	// --- 5 comp ---
	b.Run("5_comp/sorted", func(b *testing.B) {
		cursor := &view5.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view5.Filter(sortedSubset)
				for view5.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += 0.1
				}
			}
		})
	})
	b.Run("5_comp/shuffled", func(b *testing.B) {
		cursor := &view5.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view5.Filter(shuffledSubset)
				for view5.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += 0.1
				}
			}
		})
	})

	// --- 6 comp ---
	b.Run("6_comp/sorted", func(b *testing.B) {
		cursor := &view6.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view6.Filter(sortedSubset)
				for view6.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
				}
			}
		})
	})
	b.Run("6_comp/shuffled", func(b *testing.B) {
		cursor := &view6.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view6.Filter(shuffledSubset)
				for view6.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
				}
			}
		})
	})

	// --- 7 comp ---
	b.Run("7_comp/sorted", func(b *testing.B) {
		cursor := &view7.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view7.Filter(sortedSubset)
				for view7.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += 0.1
				}
			}
		})
	})
	b.Run("7_comp/shuffled", func(b *testing.B) {
		cursor := &view7.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view7.Filter(shuffledSubset)
				for view7.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += 0.1
				}
			}
		})
	})

	// --- 8 comp ---
	b.Run("8_comp/sorted", func(b *testing.B) {
		cursor := &view8.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view8.Filter(sortedSubset)
				for view8.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
				}
			}
		})
	})
	b.Run("8_comp/shuffled", func(b *testing.B) {
		cursor := &view8.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view8.Filter(shuffledSubset)
				for view8.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
				}
			}
		})
	})

	// --- 9 comp ---
	b.Run("9_comp/sorted", func(b *testing.B) {
		cursor := &view9.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view9.Filter(sortedSubset)
				for view9.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
					t09.At(cursor).V += 0.1
				}
			}
		})
	})
	b.Run("9_comp/shuffled", func(b *testing.B) {
		cursor := &view9.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view9.Filter(shuffledSubset)
				for view9.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
					t09.At(cursor).V += 0.1
				}
			}
		})
	})

	// --- 10 comp ---
	b.Run("10_comp/sorted", func(b *testing.B) {
		cursor := &view10.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view10.Filter(sortedSubset)
				for view10.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
					t09.At(cursor).V += t10.At(cursor).V
				}
			}
		})
	})
	b.Run("10_comp/shuffled", func(b *testing.B) {
		cursor := &view10.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view10.Filter(shuffledSubset)
				for view10.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
					t09.At(cursor).V += t10.At(cursor).V
				}
			}
		})
	})
}
