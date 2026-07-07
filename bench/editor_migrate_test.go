package bench_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// Benchmark_Editor_Migrate measures the per-entity cost of archetype migration
// performed one entity at a time via Editor.Update.
//
// Three scenarios, each migrating subset entities out of a population of
// entitiesNumber:
//
//	s1_addTag  {Pos}      → {Pos, Tag}   (zero-size component add)
//	s2_delTag  {Pos, Tag} → {Pos}        (zero-size component remove)
//	s3_addVel  {Pos}      → {Pos, Vel}   (data component add)
//
// After each timed iteration the counter-editor restores the original archetype
// outside the timer so the next iteration starts from the expected state.
func Benchmark_Editor_Migrate(b *testing.B) {
	ecs := setupECS()

	for _, subsetSize := range []int{filterSubsetSize, entitiesNumber} {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/s1_addTag", entitiesNumber, subsetSize), func(b *testing.B) {
			ecs.Reset()
			entities := populatePos(ecs, entitiesNumber)
			subset := entities[:subsetSize]
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag]()).Build()
			measurePerEntity(b, subsetSize, func() {
				for b.Loop() {
					for _, e := range subset {
						addEd.Update(e)
					}
					b.StopTimer()
					for _, e := range subset {
						delEd.Update(e)
					}
					b.StartTimer()
				}
			})
		})

		b.Run(fmt.Sprintf("pop=%d/subset=%d/s2_delTag", entitiesNumber, subsetSize), func(b *testing.B) {
			ecs.Reset()
			entities := populatePosTag(ecs, entitiesNumber)
			subset := entities[:subsetSize]
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag])).Build()
			measurePerEntity(b, subsetSize, func() {
				for b.Loop() {
					for _, e := range subset {
						delEd.Update(e)
					}
					b.StopTimer()
					for _, e := range subset {
						addEd.Update(e)
					}
					b.StartTimer()
				}
			})
		})

		b.Run(fmt.Sprintf("pop=%d/subset=%d/s3_addVel", entitiesNumber, subsetSize), func(b *testing.B) {
			ecs.Reset()
			entities := populatePos(ecs, entitiesNumber)
			subset := entities[:subsetSize]
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Vel])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Vel]()).Build()
			measurePerEntity(b, subsetSize, func() {
				for b.Loop() {
					for _, e := range subset {
						addEd.Update(e)
					}
					b.StopTimer()
					for _, e := range subset {
						delEd.Update(e)
					}
					b.StartTimer()
				}
			})
		})
	}
}

// Benchmark_Editor_MigrateRand is the randomly-scattered counterpart of
// Benchmark_Editor_Migrate: the migrated subset is reshuffled before every
// iteration (outside the timer), so migrated entities sit at random table
// positions instead of converging to the degenerate tail cluster that a fixed
// id list settles into.
func Benchmark_Editor_MigrateRand(b *testing.B) {
	ecs := setupECS()
	const seed = 42

	shuffle := func(r *rand.Rand, entities []uid.UID64) {
		r.Shuffle(len(entities), func(i, j int) {
			entities[i], entities[j] = entities[j], entities[i]
		})
	}

	b.Run(fmt.Sprintf("pop=%d/subset=%d/rand/s1_addTag", entitiesNumber, filterSubsetSize), func(b *testing.B) {
		ecs.Reset()
		entities := populatePos(ecs, entitiesNumber)
		addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag])).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag]()).Build()
		r := rand.New(rand.NewPCG(seed, seed))
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				b.StopTimer()
				shuffle(r, entities)
				subset := entities[:filterSubsetSize]
				b.StartTimer()
				for _, e := range subset {
					addEd.Update(e)
				}
				b.StopTimer()
				for _, e := range subset {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/subset=%d/rand/s2_delTag", entitiesNumber, filterSubsetSize), func(b *testing.B) {
		ecs.Reset()
		entities := populatePosTag(ecs, entitiesNumber)
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag]()).Build()
		addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag])).Build()
		r := rand.New(rand.NewPCG(seed, seed))
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				b.StopTimer()
				shuffle(r, entities)
				subset := entities[:filterSubsetSize]
				b.StartTimer()
				for _, e := range subset {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range subset {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/subset=%d/rand/s3_addVel", entitiesNumber, filterSubsetSize), func(b *testing.B) {
		ecs.Reset()
		entities := populatePos(ecs, entitiesNumber)
		addEd := ecs.NewEditorBuilder(new(goke.Comp[Vel])).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Vel]()).Build()
		r := rand.New(rand.NewPCG(seed, seed))
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				b.StopTimer()
				shuffle(r, entities)
				subset := entities[:filterSubsetSize]
				b.StartTimer()
				for _, e := range subset {
					addEd.Update(e)
				}
				b.StopTimer()
				for _, e := range subset {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})
}
