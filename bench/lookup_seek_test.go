package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
)

// Benchmark_Lookup_Seek measures single-entity component access via
// Lookup.Seek + Col.At, reading 0..10 component columns per entity.
//
// A Lookup resolves an entity's address directly through the index (no mask
// filtering) and bakes per-archetype column offsets lazily on the first Seek,
// caching them by archetype. The per-Seek cost is therefore dominated by the
// index lookup and cursor fill; the component count drives only how many Col.At
// reads run per entity.
//
// Same in-place-write rule as the other read benchmarks: write through the
// pointer returned by Col.At so the compiler cannot delete the store (a
// copy-then-mutate local would be eliminated and measure nothing).
//
// Lookups are created once outside b.Run: with -count=N each callback runs N
// times, so creating one per call would leak state across iterations. Because
// every lookup tracks the same components in the same prefix order, each Col[T]
// keeps a stable Idx across all lookups.
func Benchmark_Lookup_Seek(b *testing.B) {
	ecs := setupECS()
	entities := populate(ecs, entitiesNumber)
	subset := entities[:filterSubsetSize]

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

	lookup0 := goke.CreateLookup(ecs)
	lookup1 := goke.CreateLookup(ecs, goke.Track(&pos))
	lookup2 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel))
	lookup3 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc))
	lookup4 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04))
	lookup5 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05))
	lookup6 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06))
	lookup7 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07))
	lookup8 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08))
	lookup9 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09))
	lookup10 := goke.CreateLookup(ecs, goke.Track(&pos), goke.Track(&vel), goke.Track(&acc), goke.Track(&t04), goke.Track(&t05), goke.Track(&t06), goke.Track(&t07), goke.Track(&t08), goke.Track(&t09), goke.Track(&t10))

	b.Run("0 comp", func(b *testing.B) {
		var hits int
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				hits = 0
				for _, e := range subset {
					if lookup0.Seek(e) {
						hits++
					}
				}
			}
		})
		if hits != filterSubsetSize {
			b.Fatalf("Lookup sanity check failed: expected %d hits, got %d", filterSubsetSize, hits)
		}
	})

	b.Run("1 comp", func(b *testing.B) {
		cur := &lookup1.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup1.Seek(e) {
						pos.At(cur).X += pos.At(cur).Y
					}
				}
			}
		})
	})

	b.Run("2 comp", func(b *testing.B) {
		cur := &lookup2.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup2.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
					}
				}
			}
		})
	})

	b.Run("3 comp", func(b *testing.B) {
		cur := &lookup3.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup3.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += vel.At(cur).X
					}
				}
			}
		})
	})

	b.Run("4 comp", func(b *testing.B) {
		cur := &lookup4.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup4.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
					}
				}
			}
		})
	})

	b.Run("5 comp", func(b *testing.B) {
		cur := &lookup5.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup5.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += 0.1
					}
				}
			}
		})
	})

	b.Run("6 comp", func(b *testing.B) {
		cur := &lookup6.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup6.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
					}
				}
			}
		})
	})

	b.Run("7 comp", func(b *testing.B) {
		cur := &lookup7.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup7.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += 0.1
					}
				}
			}
		})
	})

	b.Run("8 comp", func(b *testing.B) {
		cur := &lookup8.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup8.Seek(e) {
						pos.At(cur).X += vel.At(cur).X
						acc.At(cur).X += t04.At(cur).V
						t05.At(cur).V += t06.At(cur).V
						t07.At(cur).V += t08.At(cur).V
					}
				}
			}
		})
	})

	b.Run("9 comp", func(b *testing.B) {
		cur := &lookup9.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup9.Seek(e) {
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

	b.Run("10 comp", func(b *testing.B) {
		cur := &lookup10.Cursor
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				for _, e := range subset {
					if lookup10.Seek(e) {
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
