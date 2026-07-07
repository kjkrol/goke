package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_Migrator_Migrate measures the per-entity cost of bulk archetype
// migration through the production path: a system iterates a Query chunk by
// chunk and registers one CmdBufMassMigrate command per chunk; Sync then
// executes Migrator.ApplyChunk for each command. Only the Sync is timed.
//
// Three scenarios, each migrating subset entities out of a population of
// entitiesNumber:
//
//	s1_addTag  {Pos}      → {Pos, Tag}   (zero-size component add)
//	s2_delTag  {Pos, Tag} → {Pos}        (zero-size component remove)
//	s3_addVel  {Pos}      → {Pos, Vel}   (data component add)
//
// subset == entitiesNumber drains whole chunks: source compaction degenerates
// to block zero + free (modeAllMigrate) with no entity relocation.
//
// After each timed Sync a counter-migration tick restores the original
// archetype outside the timer so the next iteration starts from the
// expected state.
func Benchmark_Migrator_Migrate(b *testing.B) {
	ecs := setupECS()

	for _, subset := range []int{filterSubsetSize, entitiesNumber} {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/s1_addTag", entitiesNumber, subset), func(b *testing.B) {
			ecs.Reset()
			populatePos(ecs, entitiesNumber)
			addMig := ecs.NewMigratorBuilder(new(goke.Comp[Tag])).Build()
			delMig := ecs.NewMigratorBuilder().Delete(goke.Del[Tag]()).Build()
			migrateQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Exclude(goke.Exclude[Tag]()).Build()
			restoreQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Include(goke.Include[Tag]()).Build()
			migSys := ecs.RegSysFn(enqueueSubset(migrateQ, addMig, subset))
			restoreSys := ecs.RegSysFn(enqueueAll(restoreQ, delMig))
			ecs.SetPlan(timedMigrationPlan(b, migSys, restoreSys))
			measurePerEntity(b, subset, func() {
				for b.Loop() {
					ecs.Tick(0)
				}
			})
		})

		b.Run(fmt.Sprintf("pop=%d/subset=%d/s2_delTag", entitiesNumber, subset), func(b *testing.B) {
			ecs.Reset()
			populatePosTag(ecs, entitiesNumber)
			delMig := ecs.NewMigratorBuilder().Delete(goke.Del[Tag]()).Build()
			addMig := ecs.NewMigratorBuilder(new(goke.Comp[Tag])).Build()
			migrateQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Include(goke.Include[Tag]()).Build()
			restoreQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Exclude(goke.Exclude[Tag]()).Build()
			migSys := ecs.RegSysFn(enqueueSubset(migrateQ, delMig, subset))
			restoreSys := ecs.RegSysFn(enqueueAll(restoreQ, addMig))
			ecs.SetPlan(timedMigrationPlan(b, migSys, restoreSys))
			measurePerEntity(b, subset, func() {
				for b.Loop() {
					ecs.Tick(0)
				}
			})
		})

		b.Run(fmt.Sprintf("pop=%d/subset=%d/s3_addVel", entitiesNumber, subset), func(b *testing.B) {
			ecs.Reset()
			populatePos(ecs, entitiesNumber)
			addMig := ecs.NewMigratorBuilder(new(goke.Comp[Vel])).Build()
			delMig := ecs.NewMigratorBuilder().Delete(goke.Del[Vel]()).Build()
			migrateQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Exclude(goke.Exclude[Vel]()).Build()
			restoreQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Include(goke.Include[Vel]()).Build()
			migSys := ecs.RegSysFn(enqueueSubset(migrateQ, addMig, subset))
			restoreSys := ecs.RegSysFn(enqueueAll(restoreQ, delMig))
			ecs.SetPlan(timedMigrationPlan(b, migSys, restoreSys))
			measurePerEntity(b, subset, func() {
				for b.Loop() {
					ecs.Tick(0)
				}
			})
		})
	}
}

// Benchmark_Migrator_MigrateRand is the randomly-scattered counterpart of
// Benchmark_Migrator_Migrate: the migrated subset is picked at uniformly
// random positions across the population instead of a contiguous run, so the
// source compaction takes the slot-level path instead of the block fast path.
func Benchmark_Migrator_MigrateRand(b *testing.B) {
	ecs := setupECS()
	const seed = 42

	b.Run(fmt.Sprintf("pop=%d/subset=%d/rand/s1_addTag", entitiesNumber, filterSubsetSize), func(b *testing.B) {
		ecs.Reset()
		populatePos(ecs, entitiesNumber)
		addMig := ecs.NewMigratorBuilder(new(goke.Comp[Tag])).Build()
		delMig := ecs.NewMigratorBuilder().Delete(goke.Del[Tag]()).Build()
		migrateQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Exclude(goke.Exclude[Tag]()).Build()
		restoreQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Include(goke.Include[Tag]()).Build()
		pick := randPickMask(entitiesNumber, filterSubsetSize, seed)
		migSys := ecs.RegSysFn(enqueueScattered(migrateQ, addMig, pick))
		restoreSys := ecs.RegSysFn(enqueueAll(restoreQ, delMig))
		ecs.SetPlan(timedMigrationPlan(b, migSys, restoreSys))
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				ecs.Tick(0)
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/subset=%d/rand/s2_delTag", entitiesNumber, filterSubsetSize), func(b *testing.B) {
		ecs.Reset()
		populatePosTag(ecs, entitiesNumber)
		delMig := ecs.NewMigratorBuilder().Delete(goke.Del[Tag]()).Build()
		addMig := ecs.NewMigratorBuilder(new(goke.Comp[Tag])).Build()
		migrateQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Include(goke.Include[Tag]()).Build()
		restoreQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Exclude(goke.Exclude[Tag]()).Build()
		pick := randPickMask(entitiesNumber, filterSubsetSize, seed)
		migSys := ecs.RegSysFn(enqueueScattered(migrateQ, delMig, pick))
		restoreSys := ecs.RegSysFn(enqueueAll(restoreQ, addMig))
		ecs.SetPlan(timedMigrationPlan(b, migSys, restoreSys))
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				ecs.Tick(0)
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/subset=%d/rand/s3_addVel", entitiesNumber, filterSubsetSize), func(b *testing.B) {
		ecs.Reset()
		populatePos(ecs, entitiesNumber)
		addMig := ecs.NewMigratorBuilder(new(goke.Comp[Vel])).Build()
		delMig := ecs.NewMigratorBuilder().Delete(goke.Del[Vel]()).Build()
		migrateQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Exclude(goke.Exclude[Vel]()).Build()
		restoreQ := ecs.NewQueryBuilder(new(goke.Comp[Pos])).Include(goke.Include[Vel]()).Build()
		pick := randPickMask(entitiesNumber, filterSubsetSize, seed)
		migSys := ecs.RegSysFn(enqueueScattered(migrateQ, addMig, pick))
		restoreSys := ecs.RegSysFn(enqueueAll(restoreQ, delMig))
		ecs.SetPlan(timedMigrationPlan(b, migSys, restoreSys))
		measurePerEntity(b, filterSubsetSize, func() {
			for b.Loop() {
				ecs.Tick(0)
			}
		})
	})
}
