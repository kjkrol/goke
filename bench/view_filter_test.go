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

	view0 := goke.NewView(ecs)
	view1 := goke.NewView(ecs, pos.Track())
	view2 := goke.NewView(ecs, pos.Track(), vel.Track())
	view3 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track())
	view4 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track())
	view5 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track())
	view6 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track())
	view7 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track())
	view8 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track(), t08.Track())
	view9 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track(), t08.Track(), t09.Track())
	view10 := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track(), t04.Track(), t05.Track(), t06.Track(), t07.Track(), t08.Track(), t09.Track(), t10.Track())

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
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view1.Filter(sortedSubset)
				for view1.Next() {
					pos.At(view1).X += pos.At(view1).Y
				}
			}
		})
	})
	b.Run("1_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view1.Filter(shuffledSubset)
				for view1.Next() {
					pos.At(view1).X += pos.At(view1).Y
				}
			}
		})
	})

	// --- 2 comp ---
	b.Run("2_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view2.Filter(sortedSubset)
				for view2.Next() {
					pos.At(view2).X += vel.At(view2).X
				}
			}
		})
	})
	b.Run("2_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view2.Filter(shuffledSubset)
				for view2.Next() {
					pos.At(view2).X += vel.At(view2).X
				}
			}
		})
	})

	// --- 3 comp ---
	b.Run("3_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view3.Filter(sortedSubset)
				for view3.Next() {
					pos.At(view3).X += vel.At(view3).X
					acc.At(view3).X += vel.At(view3).X
				}
			}
		})
	})
	b.Run("3_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view3.Filter(shuffledSubset)
				for view3.Next() {
					pos.At(view3).X += vel.At(view3).X
					acc.At(view3).X += vel.At(view3).X
				}
			}
		})
	})

	// --- 4 comp ---
	b.Run("4_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view4.Filter(sortedSubset)
				for view4.Next() {
					pos.At(view4).X += vel.At(view4).X
					acc.At(view4).X += t04.At(view4).V
				}
			}
		})
	})
	b.Run("4_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view4.Filter(shuffledSubset)
				for view4.Next() {
					pos.At(view4).X += vel.At(view4).X
					acc.At(view4).X += t04.At(view4).V
				}
			}
		})
	})

	// --- 5 comp ---
	b.Run("5_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view5.Filter(sortedSubset)
				for view5.Next() {
					pos.At(view5).X += vel.At(view5).X
					acc.At(view5).X += t04.At(view5).V
					t05.At(view5).V += 0.1
				}
			}
		})
	})
	b.Run("5_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view5.Filter(shuffledSubset)
				for view5.Next() {
					pos.At(view5).X += vel.At(view5).X
					acc.At(view5).X += t04.At(view5).V
					t05.At(view5).V += 0.1
				}
			}
		})
	})

	// --- 6 comp ---
	b.Run("6_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view6.Filter(sortedSubset)
				for view6.Next() {
					pos.At(view6).X += vel.At(view6).X
					acc.At(view6).X += t04.At(view6).V
					t05.At(view6).V += t06.At(view6).V
				}
			}
		})
	})
	b.Run("6_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view6.Filter(shuffledSubset)
				for view6.Next() {
					pos.At(view6).X += vel.At(view6).X
					acc.At(view6).X += t04.At(view6).V
					t05.At(view6).V += t06.At(view6).V
				}
			}
		})
	})

	// --- 7 comp ---
	b.Run("7_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view7.Filter(sortedSubset)
				for view7.Next() {
					pos.At(view7).X += vel.At(view7).X
					acc.At(view7).X += t04.At(view7).V
					t05.At(view7).V += t06.At(view7).V
					t07.At(view7).V += 0.1
				}
			}
		})
	})
	b.Run("7_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view7.Filter(shuffledSubset)
				for view7.Next() {
					pos.At(view7).X += vel.At(view7).X
					acc.At(view7).X += t04.At(view7).V
					t05.At(view7).V += t06.At(view7).V
					t07.At(view7).V += 0.1
				}
			}
		})
	})

	// --- 8 comp ---
	b.Run("8_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view8.Filter(sortedSubset)
				for view8.Next() {
					pos.At(view8).X += vel.At(view8).X
					acc.At(view8).X += t04.At(view8).V
					t05.At(view8).V += t06.At(view8).V
					t07.At(view8).V += t08.At(view8).V
				}
			}
		})
	})
	b.Run("8_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view8.Filter(shuffledSubset)
				for view8.Next() {
					pos.At(view8).X += vel.At(view8).X
					acc.At(view8).X += t04.At(view8).V
					t05.At(view8).V += t06.At(view8).V
					t07.At(view8).V += t08.At(view8).V
				}
			}
		})
	})

	// --- 9 comp ---
	b.Run("9_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view9.Filter(sortedSubset)
				for view9.Next() {
					pos.At(view9).X += vel.At(view9).X
					acc.At(view9).X += t04.At(view9).V
					t05.At(view9).V += t06.At(view9).V
					t07.At(view9).V += t08.At(view9).V
					t09.At(view9).V += 0.1
				}
			}
		})
	})
	b.Run("9_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view9.Filter(shuffledSubset)
				for view9.Next() {
					pos.At(view9).X += vel.At(view9).X
					acc.At(view9).X += t04.At(view9).V
					t05.At(view9).V += t06.At(view9).V
					t07.At(view9).V += t08.At(view9).V
					t09.At(view9).V += 0.1
				}
			}
		})
	})

	// --- 10 comp ---
	b.Run("10_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view10.Filter(sortedSubset)
				for view10.Next() {
					pos.At(view10).X += vel.At(view10).X
					acc.At(view10).X += t04.At(view10).V
					t05.At(view10).V += t06.At(view10).V
					t07.At(view10).V += t08.At(view10).V
					t09.At(view10).V += t10.At(view10).V
				}
			}
		})
	})
	b.Run("10_comp/shuffled", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				view10.Filter(shuffledSubset)
				for view10.Next() {
					pos.At(view10).X += vel.At(view10).X
					acc.At(view10).X += t04.At(view10).V
					t05.At(view10).V += t06.At(view10).V
					t07.At(view10).V += t08.At(view10).V
					t09.At(view10).V += t10.At(view10).V
				}
			}
		})
	})
}
