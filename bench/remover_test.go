package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_Remover_Remove is the bulk counterpart of Benchmark_Remove: the
// same 10-component world as the Editor suites, entities removed through
// the production CmdBuf path — a system registers one cb.Remove command per
// chunk and only the Sync executing the shared Remover's Migrate is timed.
// subset=pop removes the whole population each tick; subset<pop additionally
// exercises source compaction.
func Benchmark_Remover_Remove(b *testing.B) {
	ecs := setupECS()
	for _, subset := range []int{filterSubsetSize, entitiesNumber} {
		runRemoverLeaf(b, ecs,
			fmt.Sprintf("pop=%d/subset=%d", entitiesNumber, subset),
			subset,
			func() (*goke.Query, func()) {
				var removeQ *goke.Query
				var pop *populator
				ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
					pop = newPopulator(si)
					pop.spawn(entitiesNumber)
					removeQ = si.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
				}})
				return removeQ, func() { pop.spawn(subset) }
			})
	}
}
