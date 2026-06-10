package bench_test

import (
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

// Benchmark_View_Filter measures the performance of the View.Filter method
// across different view configurations (0 to 10 components).
// Each sub-benchmark runs the same N=100 subset under two access patterns:
//   - sorted   : entities in creation order (best case for the per-entity
//     cached archetype descriptor)
//   - shuffled : the same handles but rand.Shuffle'd (forces frequent
//     archetype-descriptor reloads)
//
// Filter yields pointers to live component memory. All variants must
// report 0 B/op and 0 allocs/op.
func Benchmark_View_Filter(b *testing.B) {
	ecs := setupECS()
	entities := populate(ecs, entitiesNumber)

	sortedSubset := append([]uid.UID64(nil), entities[:filterSubsetSize]...)
	shuffledSubset := append([]uid.UID64(nil), entities[:filterSubsetSize]...)
	rng := rand.New(rand.NewPCG(42, 1337))
	rng.Shuffle(len(shuffledSubset), func(i, j int) {
		shuffledSubset[i], shuffledSubset[j] = shuffledSubset[j], shuffledSubset[i]
	})

	// --- 0 comp ---
	b.Run("0_comp/sorted", func(b *testing.B) {
		view := goke.NewView0(ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for range view.Filter(sortedSubset) {
				}
			}
		})
	})
	b.Run("0_comp/shuffled", func(b *testing.B) {
		view := goke.NewView0(ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for range view.Filter(shuffledSubset) {
				}
			}
		})
	})

	// --- 1 comp ---
	b.Run("1_comp/sorted", func(b *testing.B) {
		view := goke.NewView1[Pos](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp1.Y
				}
			}
		})
	})
	b.Run("1_comp/shuffled", func(b *testing.B) {
		view := goke.NewView1[Pos](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp1.Y
				}
			}
		})
	})

	// --- 2 comp ---
	b.Run("2_comp/sorted", func(b *testing.B) {
		view := goke.NewView2[Pos, Vel](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
				}
			}
		})
	})
	b.Run("2_comp/shuffled", func(b *testing.B) {
		view := goke.NewView2[Pos, Vel](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
				}
			}
		})
	})

	// --- 3 comp ---
	b.Run("3_comp/sorted", func(b *testing.B) {
		view := goke.NewView3[Pos, Vel, Acc](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp2.X
				}
			}
		})
	})
	b.Run("3_comp/shuffled", func(b *testing.B) {
		view := goke.NewView3[Pos, Vel, Acc](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp2.X
				}
			}
		})
	})

	// --- 4 comp ---
	b.Run("4_comp/sorted", func(b *testing.B) {
		view := goke.NewView4[Pos, Vel, Acc, T04](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
				}
			}
		})
	})
	b.Run("4_comp/shuffled", func(b *testing.B) {
		view := goke.NewView4[Pos, Vel, Acc, T04](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
				}
			}
		})
	})

	// --- 5 comp ---
	b.Run("5_comp/sorted", func(b *testing.B) {
		view := goke.NewView5[Pos, Vel, Acc, T04, T05](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += 0.1
				}
			}
		})
	})
	b.Run("5_comp/shuffled", func(b *testing.B) {
		view := goke.NewView5[Pos, Vel, Acc, T04, T05](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += 0.1
				}
			}
		})
	})

	// --- 6 comp ---
	b.Run("6_comp/sorted", func(b *testing.B) {
		view := goke.NewView6[Pos, Vel, Acc, T04, T05, T06](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
				}
			}
		})
	})
	b.Run("6_comp/shuffled", func(b *testing.B) {
		view := goke.NewView6[Pos, Vel, Acc, T04, T05, T06](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
				}
			}
		})
	})

	// --- 7 comp ---
	b.Run("7_comp/sorted", func(b *testing.B) {
		view := goke.NewView7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += 0.1
				}
			}
		})
	})
	b.Run("7_comp/shuffled", func(b *testing.B) {
		view := goke.NewView7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += 0.1
				}
			}
		})
	})

	// --- 8 comp ---
	b.Run("8_comp/sorted", func(b *testing.B) {
		view := goke.NewView8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += item.Comp8.V
				}
			}
		})
	})
	b.Run("8_comp/shuffled", func(b *testing.B) {
		view := goke.NewView8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += item.Comp8.V
				}
			}
		})
	})

	// --- 9 comp ---
	b.Run("9_comp/sorted", func(b *testing.B) {
		view := goke.NewView9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += item.Comp8.V
					item.Comp9.V += 0.1
				}
			}
		})
	})
	b.Run("9_comp/shuffled", func(b *testing.B) {
		view := goke.NewView9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += item.Comp8.V
					item.Comp9.V += 0.1
				}
			}
		})
	})

	// --- 10 comp ---
	b.Run("10_comp/sorted", func(b *testing.B) {
		view := goke.NewView10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(sortedSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += item.Comp8.V
					item.Comp9.V += item.Comp10.V
				}
			}
		})
	})
	b.Run("10_comp/shuffled", func(b *testing.B) {
		view := goke.NewView10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, item := range view.Filter(shuffledSubset) {
					item.Comp1.X += item.Comp2.X
					item.Comp3.X += item.Comp4.V
					item.Comp5.V += item.Comp6.V
					item.Comp7.V += item.Comp8.V
					item.Comp9.V += item.Comp10.V
				}
			}
		})
	})
}
