package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_ValueMigrator_Add is the value-carrying counterpart of
// Benchmark_Migrator_Add's comps=1 leaf: identical world, population, and
// subset/order dimensions, but through ValueMigrator + CmdBufAddCompValue
// instead of Migrator + cb.Migrate — isolating the payload mechanism's
// overhead (arena reservation, per-run copyMemory, slow-path origIdx
// compaction) from everything else, since both leaves add exactly one Pos
// component onto the same Base-anchored world.
func Benchmark_ValueMigrator_Add(b *testing.B) {
	ecs := setupECS()
	for _, subset := range []int{filterSubsetSize, entitiesNumber} {
		runValueMigratorLeaf(b, ecs,
			fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=1", entitiesNumber, subset),
			subset,
			func() (*goke.Query, *goke.ValueMigrator, *goke.Comp[Pos], goke.Runnable) {
				populateBase(ecs, entitiesNumber)
				var col goke.Comp[Pos]
				fwd := ecs.NewValueMigratorBuilder(&col).Build()
				rev := ecs.NewMigratorBuilder().Remove(goke.Remove[Pos]()).Build()
				migrateQ := ecs.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Pos]()).Build()
				restoreQ := ecs.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
				return migrateQ, fwd, &col, ecs.RegSysFn(enqueueAll(restoreQ, rev))
			})
	}
}
