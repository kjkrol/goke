package bench_test

import (
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/internal/core"
)

// Benchmark_View_Filter exercises View*.Filter for every supported view
// arity (0..10 components) under two access patterns:
//
//   - sorted   : the `selected` subset is in creation order (best cache
//                locality, monotonic Link keys → chunked path skips sort)
//   - shuffled : the same handles but rand.Shuffle'd (worst case for the
//                chunked path, neutral for the per-entity inline path)
//
// Each sub-benchmark allocates its FilterCache once and reuses it across
// iterations to expose any per-call allocations the implementation might
// regress into. All variants must report 0 B/op and 0 allocs/op.
func Benchmark_View_Filter(b *testing.B) {
	ecs := setupECS()
	entities := populate(ecs, entitiesNumber)

	sortedSubset := append([]core.Entity(nil), entities[:100]...)
	shuffledSubset := append([]core.Entity(nil), entities[:100]...)
	rng := rand.New(rand.NewPCG(42, 1337))
	rng.Shuffle(len(shuffledSubset), func(i, j int) {
		shuffledSubset[i], shuffledSubset[j] = shuffledSubset[j], shuffledSubset[i]
	})

	// --- 0 comp (View0 has its own simple Filter, no FilterCache) ---
	b.Run("0_comp/sorted", func(b *testing.B) {
		view := goke.NewView0(ecs)
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for entity := range view.Filter(sortedSubset) {
					_ = entity
				}
			}
		})
	})
	b.Run("0_comp/shuffled", func(b *testing.B) {
		view := goke.NewView0(ecs)
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for entity := range view.Filter(shuffledSubset) {
					_ = entity
				}
			}
		})
	})

	// --- 1 comp ---
	b.Run("1_comp/sorted", func(b *testing.B) {
		view := goke.NewView1[Pos](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						pos := &page.Comp1[i]
						pos.X += pos.Y
					}
				}
			}
		})
	})
	b.Run("1_comp/shuffled", func(b *testing.B) {
		view := goke.NewView1[Pos](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						pos := &page.Comp1[i]
						pos.X += pos.Y
					}
				}
			}
		})
	})

	// --- 2 comp ---
	b.Run("2_comp/sorted", func(b *testing.B) {
		view := goke.NewView2[Pos, Vel](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
					}
				}
			}
		})
	})
	b.Run("2_comp/shuffled", func(b *testing.B) {
		view := goke.NewView2[Pos, Vel](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
					}
				}
			}
		})
	})

	// --- 3 comp ---
	b.Run("3_comp/sorted", func(b *testing.B) {
		view := goke.NewView3[Pos, Vel, Acc](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp2[i].X
					}
				}
			}
		})
	})
	b.Run("3_comp/shuffled", func(b *testing.B) {
		view := goke.NewView3[Pos, Vel, Acc](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp2[i].X
					}
				}
			}
		})
	})

	// --- 4 comp ---
	b.Run("4_comp/sorted", func(b *testing.B) {
		view := goke.NewView4[Pos, Vel, Acc, T04](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
					}
				}
			}
		})
	})
	b.Run("4_comp/shuffled", func(b *testing.B) {
		view := goke.NewView4[Pos, Vel, Acc, T04](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
					}
				}
			}
		})
	})

	// --- 5 comp ---
	b.Run("5_comp/sorted", func(b *testing.B) {
		view := goke.NewView5[Pos, Vel, Acc, T04, T05](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += 0.1
					}
				}
			}
		})
	})
	b.Run("5_comp/shuffled", func(b *testing.B) {
		view := goke.NewView5[Pos, Vel, Acc, T04, T05](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += 0.1
					}
				}
			}
		})
	})

	// --- 6 comp ---
	b.Run("6_comp/sorted", func(b *testing.B) {
		view := goke.NewView6[Pos, Vel, Acc, T04, T05, T06](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
					}
				}
			}
		})
	})
	b.Run("6_comp/shuffled", func(b *testing.B) {
		view := goke.NewView6[Pos, Vel, Acc, T04, T05, T06](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
					}
				}
			}
		})
	})

	// --- 7 comp ---
	b.Run("7_comp/sorted", func(b *testing.B) {
		view := goke.NewView7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += 0.1
					}
				}
			}
		})
	})
	b.Run("7_comp/shuffled", func(b *testing.B) {
		view := goke.NewView7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += 0.1
					}
				}
			}
		})
	})

	// --- 8 comp ---
	b.Run("8_comp/sorted", func(b *testing.B) {
		view := goke.NewView8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += page.Comp8[i].V
					}
				}
			}
		})
	})
	b.Run("8_comp/shuffled", func(b *testing.B) {
		view := goke.NewView8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += page.Comp8[i].V
					}
				}
			}
		})
	})

	// --- 9 comp ---
	b.Run("9_comp/sorted", func(b *testing.B) {
		view := goke.NewView9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += page.Comp8[i].V
						page.Comp9[i].V += 0.1
					}
				}
			}
		})
	})
	b.Run("9_comp/shuffled", func(b *testing.B) {
		view := goke.NewView9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += page.Comp8[i].V
						page.Comp9[i].V += 0.1
					}
				}
			}
		})
	})

	// --- 10 comp ---
	b.Run("10_comp/sorted", func(b *testing.B) {
		view := goke.NewView10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		var cache goke.FilterCache
		cache.Grow(len(sortedSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(sortedSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += page.Comp8[i].V
						page.Comp9[i].V += page.Comp10[i].V
					}
				}
			}
		})
	})
	b.Run("10_comp/shuffled", func(b *testing.B) {
		view := goke.NewView10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		var cache goke.FilterCache
		cache.Grow(len(shuffledSubset))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for page := range view.Filter(shuffledSubset, &cache) {
					for i := range page.Entity {
						page.Comp1[i].X += page.Comp2[i].X
						page.Comp3[i].X += page.Comp4[i].V
						page.Comp5[i].V += page.Comp6[i].V
						page.Comp7[i].V += page.Comp8[i].V
						page.Comp9[i].V += page.Comp10[i].V
					}
				}
			}
		})
	})
}
