package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_Editor_Add exercises a structural add through the production
// CmdBuf path: the same worlds and payloads (comps=N data components or
// tags=N zero-size tags added onto a Base anchor), but a system registers
// one cb.Migrate command per chunk and only the Sync executing
// Editor.Migrate is timed. subset<pop leaves additionally exercise source
// compaction.
func Benchmark_Editor_Add(b *testing.B) {
	ecs := setupECS()
	addComps := []goke.Addable{
		new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]),
		new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]),
		new(goke.Comp[T07]), new(goke.Comp[T08]), new(goke.Comp[T09]),
		new(goke.Comp[T10]),
	}
	delComps := []goke.EditOpt{
		goke.Remove[Pos](), goke.Remove[Vel](), goke.Remove[Acc](),
		goke.Remove[T04](), goke.Remove[T05](), goke.Remove[T06](),
		goke.Remove[T07](), goke.Remove[T08](), goke.Remove[T09](),
		goke.Remove[T10](),
	}

	for n := 1; n <= 10; n++ {
		for _, subset := range []int{filterSubsetSize, entitiesNumber} {
			runEditorLeaf(b, ecs,
				fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=%d", entitiesNumber, subset, n),
				subset,
				func() (*goke.Query, *goke.Editor, goke.Runnable) {
					var migrateQ, restoreQ *goke.Query
					var fwd, rev *goke.Editor
					ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
						populateBase(si, entitiesNumber)
						migrateQ = si.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Pos]()).Build()
						restoreQ = si.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
						fwd = migrateQ.NewEditorBuilder(addComps[:n]...).Build()
						rev = restoreQ.NewEditorBuilder().Remove(delComps[:n]...).Build()
					}})
					return migrateQ, fwd, ecs.RegSys(enqueueAll(restoreQ, rev))
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
		goke.Remove[Tag1](), goke.Remove[Tag2](), goke.Remove[Tag3](),
		goke.Remove[Tag4](), goke.Remove[Tag5](), goke.Remove[Tag6](),
		goke.Remove[Tag7](), goke.Remove[Tag8](), goke.Remove[Tag9](),
		goke.Remove[Tag10](),
	}

	for n := 1; n <= 10; n++ {
		for _, subset := range []int{filterSubsetSize, entitiesNumber} {
			runEditorLeaf(b, tagECS,
				fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=%d", entitiesNumber, subset, n),
				subset,
				func() (*goke.Query, *goke.Editor, goke.Runnable) {
					var migrateQ, restoreQ *goke.Query
					var fwd, rev *goke.Editor
					tagECS.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
						populateBase(si, entitiesNumber)
						migrateQ = si.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Tag1]()).Build()
						restoreQ = si.NewQueryBuilder().Include(goke.Include[Tag1]()).Build()
						fwd = migrateQ.NewEditorBuilder(addTags[:n]...).Build()
						rev = restoreQ.NewEditorBuilder().Remove(delTags[:n]...).Build()
					}})
					return migrateQ, fwd, tagECS.RegSys(enqueueAll(restoreQ, rev))
				})
		}
	}
}
