package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_Migrator_Add is the bulk counterpart of Benchmark_Editor_Add:
// the same worlds and payloads (comps=N data components or tags=N zero-size
// tags added onto a Base anchor), but migrated through the production CmdBuf
// path — a system registers one CmdBufMassMigrate command per chunk and only
// the Sync executing Migrator.Migrate is timed. subset=pop leaves pair 1:1
// with Editor_Add; subset<pop leaves additionally exercise source compaction.
func Benchmark_Migrator_Add(b *testing.B) {
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
				fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=%d", entitiesNumber, subset, n),
				subset,
				func() (*goke.Query, *goke.Migrator, goke.System) {
					populateBase(ecs, entitiesNumber)
					fwd := ecs.NewMigratorBuilder(addComps[:n]...).Build()
					rev := ecs.NewMigratorBuilder().Delete(delComps[:n]...).Build()
					migrateQ := ecs.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Pos]()).Build()
					restoreQ := ecs.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
					return migrateQ, fwd, ecs.RegSysFn(enqueueAll(restoreQ, rev))
				})
		}
	}

	tagECS := setupTagECS()
	addTags := []goke.Addable{
		new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]),
		new(goke.Comp[Tag4]), new(goke.Comp[Tag5]), new(goke.Comp[Tag6]),
		new(goke.Comp[Tag7]), new(goke.Comp[Tag8]), new(goke.Comp[Tag9]),
		new(goke.Comp[Tag10]),
	}
	delTags := []goke.EditOpt{
		goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](),
		goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](),
		goke.Del[Tag7](), goke.Del[Tag8](), goke.Del[Tag9](),
		goke.Del[Tag10](),
	}

	for n := 1; n <= 10; n++ {
		for _, subset := range []int{filterSubsetSize, entitiesNumber} {
			runMigratorLeaf(b, tagECS,
				fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=%d", entitiesNumber, subset, n),
				subset,
				func() (*goke.Query, *goke.Migrator, goke.System) {
					populateBase(tagECS, entitiesNumber)
					fwd := tagECS.NewMigratorBuilder(addTags[:n]...).Build()
					rev := tagECS.NewMigratorBuilder().Delete(delTags[:n]...).Build()
					migrateQ := tagECS.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Tag1]()).Build()
					restoreQ := tagECS.NewQueryBuilder().Include(goke.Include[Tag1]()).Build()
					return migrateQ, fwd, tagECS.RegSysFn(enqueueAll(restoreQ, rev))
				})
		}
	}
}
