package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_ValueEditor_Add is the value-carrying counterpart of
// Benchmark_Editor_Add's comps=1 leaf: identical world, population, and
// subset/order dimensions, but through ValueEditor + CmdBufAddCompValue
// instead of Editor + cb.Migrate — isolating the payload mechanism's
// overhead (arena reservation, per-run copyMemory, slow-path origIdx
// compaction) from everything else, since both leaves add exactly one Pos
// component onto the same Base-anchored world.
func Benchmark_ValueEditor_Add(b *testing.B) {
	ecs := setupECS()
	for _, subset := range []int{filterSubsetSize, entitiesNumber} {
		runValueEditorLeaf(b, ecs,
			fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=1", entitiesNumber, subset),
			subset,
			func() (*goke.Query, *goke.ValueEditor, *goke.Comp[Pos], goke.Runnable) {
				var col goke.Comp[Pos]
				var migrateQ, restoreQ *goke.Query
				var fwd *goke.ValueEditor
				var rev *goke.Editor
				ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
					populateBase(si, entitiesNumber)
					migrateQ = si.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Pos]()).Build()
					restoreQ = si.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
					fwd = migrateQ.NewValueEditorBuilder(&col).Build()
					rev = restoreQ.NewEditorBuilder().Remove(goke.Remove[Pos]()).Build()
				}})
				return migrateQ, fwd, &col, ecs.RegSys(enqueueAll(restoreQ, rev))
			})
	}
}
