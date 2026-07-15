package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_Migrator_Del is the bulk counterpart of Benchmark_Editor_Del:
// entities carry 10 data components and each leaf removes comps=N of them
// through the production CmdBuf path; only the Sync executing
// Migrator.Migrate is timed. Like Editor_Del, the comps=10 world anchors
// entities with an extra Base component so removing all 10 tracked
// components never unlinks them. subset=pop leaves pair 1:1 with
// Editor_Del; subset<pop leaves additionally exercise source compaction.
func Benchmark_Migrator_Del(b *testing.B) {
	ecs := setupECS()
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

	for n := 1; n <= 10; n++ {
		for _, subset := range []int{filterSubsetSize, entitiesNumber} {
			runMigratorLeaf(b, ecs,
				fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=%d", entitiesNumber, subset, n),
				subset,
				func() (*goke.Query, *goke.Migrator, goke.System) {
					entities := populate(ecs, entitiesNumber)
					if n == 10 {
						anchorEd := ecs.NewEditorBuilder(new(goke.Comp[Base])).Build()
						for _, e := range entities {
							anchorEd.Update(e)
						}
					}
					fwd := ecs.NewMigratorBuilder().Delete(delComps[:n]...).Build()
					rev := ecs.NewMigratorBuilder(addComps[:n]...).Build()
					migrateQ := ecs.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
					var restoreQ *goke.Query
					if n == 10 {
						restoreQ = ecs.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Pos]()).Build()
					} else {
						restoreQ = ecs.NewQueryBuilder().Include(goke.Include[T10]()).Exclude(goke.Exclude[Pos]()).Build()
					}
					return migrateQ, fwd, ecs.RegSysFn(enqueueAll(restoreQ, rev))
				})
		}
	}
}
