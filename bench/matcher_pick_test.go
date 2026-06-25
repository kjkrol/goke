package bench_test

import (
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// Benchmark_Matcher_Pick measures the performance of the Matcher.Pick method
// across different query configurations (0 to 10 components), drawing a
// subset of filterSubsetSize entities from a population of entitiesNumber
// under two access patterns:
//   - sorted : the first filterSubsetSize entities in creation order (best
//     case for the per-entityID cached archetype descriptor)
//   - random : filterSubsetSize entities randomly sampled from the full
//     population, in random order (forces frequent archetype-descriptor
//     reloads and exercises cache-unfriendly access)
//
// Pick yields pointers to live component memory. All variants must
// report 0 B/op and 0 allocs/op.
//
// Matchers are created once outside b.Run: with -count=N each b.Run callback is
// called N times, so creating a new Matcher inside would accumulate N matchers per
// sub-benchmark on the same ECS and eventually exceed MaxMatchers.
func Benchmark_Matcher_Pick(b *testing.B) {
	ecs := setupECS()
	entities := populate(ecs, entitiesNumber)

	sortedSubset := append([]uid.UID64(nil), entities[:filterSubsetSize]...)

	randomSubset := append([]uid.UID64(nil), entities...)
	rng := rand.New(rand.NewPCG(42, 1337))
	rng.Shuffle(len(randomSubset), func(i, j int) {
		randomSubset[i], randomSubset[j] = randomSubset[j], randomSubset[i]
	})
	randomSubset = randomSubset[:filterSubsetSize]

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
	matcher3 := ecs.NewQueryBuilder(&pos, &vel, &acc).Build()
	matcher4 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04).Build()
	matcher5 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05).Build()
	matcher6 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06).Build()
	matcher7 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07).Build()
	matcher8 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08).Build()
	matcher9 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09).Build()
	matcher10 := ecs.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09, &t10).Build()

	// --- 0 comp ---
	b.Run("pop=1024/subset=100/0_comp/sorted", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher0.Pick(sortedSubset)
				for matcher0.Next() {
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/0_comp/random", func(b *testing.B) {
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher0.Pick(randomSubset)
				for matcher0.Next() {
				}
			}
		})
	})

	// --- 1 comp ---
	b.Run("pop=1024/subset=100/1_comp/sorted", func(b *testing.B) {
		cursor := matcher1.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher1.Pick(sortedSubset)
				for matcher1.Next() {
					pos.At(cursor).X += pos.At(cursor).Y
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/1_comp/random", func(b *testing.B) {
		cursor := matcher1.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher1.Pick(randomSubset)
				for matcher1.Next() {
					pos.At(cursor).X += pos.At(cursor).Y
				}
			}
		})
	})

	// --- 2 comp ---
	b.Run("pop=1024/subset=100/2_comp/sorted", func(b *testing.B) {
		cursor := matcher2.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher2.Pick(sortedSubset)
				for matcher2.Next() {
					pos.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/2_comp/random", func(b *testing.B) {
		cursor := matcher2.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher2.Pick(randomSubset)
				for matcher2.Next() {
					pos.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})

	// --- 3 comp ---
	b.Run("pop=1024/subset=100/3_comp/sorted", func(b *testing.B) {
		cursor := matcher3.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher3.Pick(sortedSubset)
				for matcher3.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/3_comp/random", func(b *testing.B) {
		cursor := matcher3.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher3.Pick(randomSubset)
				for matcher3.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += vel.At(cursor).X
				}
			}
		})
	})

	// --- 4 comp ---
	b.Run("pop=1024/subset=100/4_comp/sorted", func(b *testing.B) {
		cursor := matcher4.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher4.Pick(sortedSubset)
				for matcher4.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/4_comp/random", func(b *testing.B) {
		cursor := matcher4.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher4.Pick(randomSubset)
				for matcher4.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
				}
			}
		})
	})

	// --- 5 comp ---
	b.Run("pop=1024/subset=100/5_comp/sorted", func(b *testing.B) {
		cursor := matcher5.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher5.Pick(sortedSubset)
				for matcher5.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += 0.1
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/5_comp/random", func(b *testing.B) {
		cursor := matcher5.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher5.Pick(randomSubset)
				for matcher5.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += 0.1
				}
			}
		})
	})

	// --- 6 comp ---
	b.Run("pop=1024/subset=100/6_comp/sorted", func(b *testing.B) {
		cursor := matcher6.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher6.Pick(sortedSubset)
				for matcher6.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/6_comp/random", func(b *testing.B) {
		cursor := matcher6.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher6.Pick(randomSubset)
				for matcher6.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
				}
			}
		})
	})

	// --- 7 comp ---
	b.Run("pop=1024/subset=100/7_comp/sorted", func(b *testing.B) {
		cursor := matcher7.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher7.Pick(sortedSubset)
				for matcher7.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += 0.1
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/7_comp/random", func(b *testing.B) {
		cursor := matcher7.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher7.Pick(randomSubset)
				for matcher7.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += 0.1
				}
			}
		})
	})

	// --- 8 comp ---
	b.Run("pop=1024/subset=100/8_comp/sorted", func(b *testing.B) {
		cursor := matcher8.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher8.Pick(sortedSubset)
				for matcher8.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/8_comp/random", func(b *testing.B) {
		cursor := matcher8.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher8.Pick(randomSubset)
				for matcher8.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
				}
			}
		})
	})

	// --- 9 comp ---
	b.Run("pop=1024/subset=100/9_comp/sorted", func(b *testing.B) {
		cursor := matcher9.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher9.Pick(sortedSubset)
				for matcher9.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
					t09.At(cursor).V += 0.1
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/9_comp/random", func(b *testing.B) {
		cursor := matcher9.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher9.Pick(randomSubset)
				for matcher9.Next() {
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
	b.Run("pop=1024/subset=100/10_comp/sorted", func(b *testing.B) {
		cursor := matcher10.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher10.Pick(sortedSubset)
				for matcher10.Next() {
					pos.At(cursor).X += vel.At(cursor).X
					acc.At(cursor).X += t04.At(cursor).V
					t05.At(cursor).V += t06.At(cursor).V
					t07.At(cursor).V += t08.At(cursor).V
					t09.At(cursor).V += t10.At(cursor).V
				}
			}
		})
	})
	b.Run("pop=1024/subset=100/10_comp/random", func(b *testing.B) {
		cursor := matcher10.Cursor()
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				matcher10.Pick(randomSubset)
				for matcher10.Next() {
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
