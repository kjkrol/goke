package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_Migrator_Mix is the bulk counterpart of Benchmark_Editor_Mix:
// each leaf swap=N adds N new types (E01..E0N) and removes N existing ones
// (Pos..) in a single bulk migration through the production CmdBuf path;
// only the Sync executing Migrator.Migrate is timed. subset=pop leaves pair
// 1:1 with Editor_Mix; subset<pop leaves additionally exercise source
// compaction.
func Benchmark_Migrator_Mix(b *testing.B) {
	ecs := setupEditECS()
	addComps := []goke.Addable{
		new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]),
		new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]),
		new(goke.Comp[T07]), new(goke.Comp[T08]), new(goke.Comp[T09]),
		new(goke.Comp[T10]),
	}
	delComps := []goke.EditOpt{
		goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
		goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
		goke.Del[T07](), goke.Del[T08](), goke.Del[T09](),
		goke.Del[T10](),
	}
	addE := []goke.Addable{
		new(goke.Comp[E01]), new(goke.Comp[E02]), new(goke.Comp[E03]),
		new(goke.Comp[E04]), new(goke.Comp[E05]), new(goke.Comp[E06]),
		new(goke.Comp[E07]), new(goke.Comp[E08]), new(goke.Comp[E09]),
		new(goke.Comp[E10]),
	}
	delE := []goke.EditOpt{
		goke.Del[E01](), goke.Del[E02](), goke.Del[E03](),
		goke.Del[E04](), goke.Del[E05](), goke.Del[E06](),
		goke.Del[E07](), goke.Del[E08](), goke.Del[E09](),
		goke.Del[E10](),
	}

	for n := 1; n <= 10; n++ {
		for _, subset := range []int{filterSubsetSize, entitiesNumber} {
			runMigratorLeaf(b, ecs,
				fmt.Sprintf("pop=%d/subset=%d/comp=%d/swap=%d", entitiesNumber, subset, n+1, n),
				subset,
				func() (*goke.Query, *goke.Migrator, goke.System) {
					populateN(ecs, entitiesNumber, n)
					fwd := ecs.NewMigratorBuilder(addE[:n]...).Delete(delComps[:n]...).Build()
					rev := ecs.NewMigratorBuilder(addComps[:n]...).Delete(delE[:n]...).Build()
					migrateQ := ecs.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
					restoreQ := ecs.NewQueryBuilder().Include(goke.Include[E01]()).Build()
					return migrateQ, fwd, ecs.RegSysFn(enqueueAll(restoreQ, rev))
				})
		}
	}
}
