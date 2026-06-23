package bench_test

import (
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

// Benchmark_Matcher_Seek measures single-entity component access via
// Matcher.Seek + Col.At, reading 0..10 component columns per entity, over a
// subset of filterSubsetSize entities drawn from a population of
// entitiesNumber, under two access patterns:
//   - (default) : the first filterSubsetSize entities in creation order
//   - random    : filterSubsetSize entities randomly sampled from the full
//     population, in random order (cache-unfriendly, jumps between
//     archetype/chunk locations)
//
// A Matcher.Seek resolves an entity's address directly through the index (no mask
// filtering) and bakes per-archetype column offsets lazily on the first Seek,
// caching them by archetype. The per-Seek cost is therefore dominated by the
// index lookup and cursor fill; the component count drives only how many Col.At
// reads run per entity.
//
// Same in-place-write rule as the other read benchmarks: write through the
// pointer returned by Col.At so the compiler cannot delete the store (a
// copy-then-mutate local would be eliminated and measure nothing).
//
// Matchers are created once outside b.Run: with -count=N each callback runs N
// times, so creating one per call would leak state across iterations. Because
// every matcher tracks the same components in the same prefix order, each Col[T]
// keeps a stable Idx across all matchers.
func Benchmark_Matcher_Seek(b *testing.B) {
	ecs := setupECS()
	entities := populate(ecs, entitiesNumber)
	subset := entities[:filterSubsetSize]

	randomSubset := append([]uid.UID64(nil), entities...)
	rng := rand.New(rand.NewPCG(42, 1337))
	rng.Shuffle(len(randomSubset), func(i, j int) {
		randomSubset[i], randomSubset[j] = randomSubset[j], randomSubset[i]
	})
	randomSubset = randomSubset[:filterSubsetSize]

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

	matcher0 := ecs.CreateMatcher()
	matcher1 := ecs.CreateMatcher(goke.Track(&pos))
	matcher2 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel))
	matcher3 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc))
	matcher4 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04))
	matcher5 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05))
	matcher6 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06))
	matcher7 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07))
	matcher8 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08))
	matcher9 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09))
	matcher10 := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09), goke.Track(&t10))

	b.Run("pop=1024/0_comp", func(b *testing.B) {
		var hits int
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				hits = 0
				for _, e := range subset {
					if matcher0.Seek(e) {
						hits++
					}
				}
			}
		})
		if hits != filterSubsetSize {
			b.Fatalf("Matcher sanity check failed: expected %d hits, got %d", filterSubsetSize, hits)
		}
	})
	b.Run("pop=1024/0_comp/random", func(b *testing.B) {
		var hits int
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				hits = 0
				for _, e := range randomSubset {
					if matcher0.Seek(e) {
						hits++
					}
				}
			}
		})
		if hits != filterSubsetSize {
			b.Fatalf("Matcher sanity check failed: expected %d hits, got %d", filterSubsetSize, hits)
		}
	})

	b.Run("pop=1024/1_comp", func(b *testing.B) {
		cur := &matcher1.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher1.Seek(e) {
						pos.At(cur).X += pos.At(cur).Y
					}
				}
			}
		})
	})
	b.Run("pop=1024/1_comp/random", func(b *testing.B) {
		cur := &matcher1.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher1.Seek(e) {
						pos.At(cur).X += pos.At(cur).Y
					}
				}
			}
		})
	})

	b.Run("pop=1024/2_comp", func(b *testing.B) {
		cur := &matcher2.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher2.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
					}
				}
			}
		})
	})
	b.Run("pop=1024/2_comp/random", func(b *testing.B) {
		cur := &matcher2.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher2.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
					}
				}
			}
		})
	})

	b.Run("pop=1024/3_comp", func(b *testing.B) {
		cur := &matcher3.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher3.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += vel.At(cur).X
					}
				}
			}
		})
	})
	b.Run("pop=1024/3_comp/random", func(b *testing.B) {
		cur := &matcher3.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher3.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += vel.At(cur).X
					}
				}
			}
		})
	})

	b.Run("pop=1024/4_comp", func(b *testing.B) {
		cur := &matcher4.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher4.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
					}
				}
			}
		})
	})
	b.Run("pop=1024/4_comp/random", func(b *testing.B) {
		cur := &matcher4.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher4.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
					}
				}
			}
		})
	})

	b.Run("pop=1024/5_comp", func(b *testing.B) {
		cur := &matcher5.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher5.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += 0.1
					}
				}
			}
		})
	})
	b.Run("pop=1024/5_comp/random", func(b *testing.B) {
		cur := &matcher5.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher5.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += 0.1
					}
				}
			}
		})
	})

	b.Run("pop=1024/6_comp", func(b *testing.B) {
		cur := &matcher6.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher6.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
					}
				}
			}
		})
	})
	b.Run("pop=1024/6_comp/random", func(b *testing.B) {
		cur := &matcher6.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher6.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
					}
				}
			}
		})
	})

	b.Run("pop=1024/7_comp", func(b *testing.B) {
		cur := &matcher7.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher7.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += 0.1
					}
				}
			}
		})
	})
	b.Run("pop=1024/7_comp/random", func(b *testing.B) {
		cur := &matcher7.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher7.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += 0.1
					}
				}
			}
		})
	})

	b.Run("pop=1024/8_comp", func(b *testing.B) {
		cur := &matcher8.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher8.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += t08.At(cur).V
					}
				}
			}
		})
	})
	b.Run("pop=1024/8_comp/random", func(b *testing.B) {
		cur := &matcher8.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher8.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += t08.At(cur).V
					}
				}
			}
		})
	})

	b.Run("pop=1024/9_comp", func(b *testing.B) {
		cur := &matcher9.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher9.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += t08.At(cur).V
						t09.At(cur).V += 0.1
					}
				}
			}
		})
	})
	b.Run("pop=1024/9_comp/random", func(b *testing.B) {
		cur := &matcher9.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher9.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += t08.At(cur).V
						t09.At(cur).V += 0.1
					}
				}
			}
		})
	})

	b.Run("pop=1024/10_comp", func(b *testing.B) {
		cur := &matcher10.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if matcher10.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += t08.At(cur).V
						t09.At(cur).V += t10.At(cur).V
					}
				}
			}
		})
	})
	b.Run("pop=1024/10_comp/random", func(b *testing.B) {
		cur := &matcher10.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range randomSubset {
					if matcher10.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += t08.At(cur).V
						t09.At(cur).V += t10.At(cur).V
					}
				}
			}
		})
	})
}
