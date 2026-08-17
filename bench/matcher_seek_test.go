package bench_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// Benchmark_Matcher_Seek measures single-entity component access via
// Query.Seek + Comp.At, reading 0..10 component columns per entity, over a
// subset of filterSubsetSize entities drawn from a population of
// entitiesNumber. sorted seeks the first filterSubsetSize entities in
// creation order; random seeks a random sample of the full population in
// random order (cache-unfriendly index jumps).
//
// Writes go through the pointer returned by Comp.At so the compiler cannot
// delete the store. Matchers are created once outside b.Run: with -count=N
// each callback runs N times, so creating one per call would leak state
// across iterations.
func Benchmark_Matcher_Seek(b *testing.B) {
	ecs := setupECS()

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

	var entities []uid.UID64
	var matcher0, matcher1, matcher2, matcher3, matcher4, matcher5, matcher6, matcher7, matcher8, matcher9, matcher10 *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		entities = populate(si, entitiesNumber)
		matcher0 = si.NewQueryBuilder().Build()
		matcher1 = si.NewQueryBuilder(&pos).Build()
		matcher2 = si.NewQueryBuilder(&pos, &vel).Build()
		matcher3 = si.NewQueryBuilder(&pos, &vel, &acc).Build()
		matcher4 = si.NewQueryBuilder(&pos, &vel, &acc, &t04).Build()
		matcher5 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05).Build()
		matcher6 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06).Build()
		matcher7 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07).Build()
		matcher8 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08).Build()
		matcher9 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09).Build()
		matcher10 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09, &t10).Build()
	}})

	sortedIDs := entities[:filterSubsetSize]
	randomIDs := append([]uid.UID64(nil), entities...)
	rng := rand.New(rand.NewPCG(42, 1337))
	rng.Shuffle(len(randomIDs), func(i, j int) {
		randomIDs[i], randomIDs[j] = randomIDs[j], randomIDs[i]
	})
	randomIDs = randomIDs[:filterSubsetSize]
	orders := []struct {
		name string
		ids  []uid.UID64
	}{{"sorted", sortedIDs}, {"random", randomIDs}}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=0/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			var hits int
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					hits = 0
					for _, e := range ids {
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
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher1.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
						if matcher1.Seek(e) {
							pos.At(cur).X += pos.At(cur).Y
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=2/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher2.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
						if matcher2.Seek(e) {
							pos.At(cur).X += vel.At(cur).X
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=3/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher3.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
						if matcher3.Seek(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += vel.At(cur).X
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=4/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher4.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
						if matcher4.Seek(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=5/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher5.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
						if matcher5.Seek(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
							t05.At(cur).V += 0.1
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=6/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher6.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
						if matcher6.Seek(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
							t05.At(cur).V += t06.At(cur).V
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=7/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher7.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
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
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=8/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher8.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
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
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=9/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher9.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
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
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher10.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					for _, e := range ids {
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
}

// Benchmark_Matcher_SeekH is the SeekH counterpart of Benchmark_Matcher_Seek.
// SeekH assumes the entity is alive and shares the archetype cached by a
// prior Seek, skipping the generation and archetype-change checks; every
// entity here lives in the one populated archetype, so each iteration seeds
// the cache with one Seek and runs SeekH for the whole subset.
func Benchmark_Matcher_SeekH(b *testing.B) {
	ecs := setupECS()

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

	var entities []uid.UID64
	var matcher0, matcher1, matcher2, matcher3, matcher4, matcher5, matcher6, matcher7, matcher8, matcher9, matcher10 *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		entities = populate(si, entitiesNumber)
		matcher0 = si.NewQueryBuilder().Build()
		matcher1 = si.NewQueryBuilder(&pos).Build()
		matcher2 = si.NewQueryBuilder(&pos, &vel).Build()
		matcher3 = si.NewQueryBuilder(&pos, &vel, &acc).Build()
		matcher4 = si.NewQueryBuilder(&pos, &vel, &acc, &t04).Build()
		matcher5 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05).Build()
		matcher6 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06).Build()
		matcher7 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07).Build()
		matcher8 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08).Build()
		matcher9 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09).Build()
		matcher10 = si.NewQueryBuilder(&pos, &vel, &acc, &t04, &t05, &t06, &t07, &t08, &t09, &t10).Build()
	}})

	sortedIDs := entities[:filterSubsetSize]
	randomIDs := append([]uid.UID64(nil), entities...)
	rng := rand.New(rand.NewPCG(42, 1337))
	rng.Shuffle(len(randomIDs), func(i, j int) {
		randomIDs[i], randomIDs[j] = randomIDs[j], randomIDs[i]
	})
	randomIDs = randomIDs[:filterSubsetSize]
	orders := []struct {
		name string
		ids  []uid.UID64
	}{{"sorted", sortedIDs}, {"random", randomIDs}}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=0/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			var hits int
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					hits = 0
					matcher0.Seek(ids[0])
					for _, e := range ids {
						if matcher0.SeekH(e) {
							hits++
						}
					}
				}
			})
			if hits != filterSubsetSize {
				b.Fatalf("Matcher sanity check failed: expected %d hits, got %d", filterSubsetSize, hits)
			}
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher1.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher1.Seek(ids[0])
					for _, e := range ids {
						if matcher1.SeekH(e) {
							pos.At(cur).X += pos.At(cur).Y
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=2/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher2.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher2.Seek(ids[0])
					for _, e := range ids {
						if matcher2.SeekH(e) {
							pos.At(cur).X += vel.At(cur).X
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=3/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher3.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher3.Seek(ids[0])
					for _, e := range ids {
						if matcher3.SeekH(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += vel.At(cur).X
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=4/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher4.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher4.Seek(ids[0])
					for _, e := range ids {
						if matcher4.SeekH(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=5/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher5.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher5.Seek(ids[0])
					for _, e := range ids {
						if matcher5.SeekH(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
							t05.At(cur).V += 0.1
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=6/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher6.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher6.Seek(ids[0])
					for _, e := range ids {
						if matcher6.SeekH(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
							t05.At(cur).V += t06.At(cur).V
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=7/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher7.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher7.Seek(ids[0])
					for _, e := range ids {
						if matcher7.SeekH(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
							t05.At(cur).V += t06.At(cur).V
							t07.At(cur).V += 0.1
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=8/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher8.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher8.Seek(ids[0])
					for _, e := range ids {
						if matcher8.SeekH(e) {
							pos.At(cur).X += vel.At(cur).X
							acc.At(cur).X += t04.At(cur).V
							t05.At(cur).V += t06.At(cur).V
							t07.At(cur).V += t08.At(cur).V
						}
					}
				}
			})
		})
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=9/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher9.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher9.Seek(ids[0])
					for _, e := range ids {
						if matcher9.SeekH(e) {
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
	}

	for _, o := range orders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/%s", entitiesNumber, filterSubsetSize, o.name), func(b *testing.B) {
			ids := o.ids
			cur := matcher10.Cursor()
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					matcher10.Seek(ids[0])
					for _, e := range ids {
						if matcher10.SeekH(e) {
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
}
