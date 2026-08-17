package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// populateWithBase mirrors populate but additionally spawns each entity with
// a Base anchor in the same Factory call, so removing all 10 tracked
// components in the n==10 case never unlinks the entity — no retrofit pass
// needed after the fact.
func populateWithBase(si *goke.SysInit, count int) []uid.UID64 {
	var cBase goke.Comp[Base]
	var c1 goke.Comp[Pos]
	var c2 goke.Comp[Vel]
	var c3 goke.Comp[Acc]
	var c4 goke.Comp[T04]
	var c5 goke.Comp[T05]
	var c6 goke.Comp[T06]
	var c7 goke.Comp[T07]
	var c8 goke.Comp[T08]
	var c9 goke.Comp[T09]
	var c10 goke.Comp[T10]
	factory := si.NewFactory(&cBase, &c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10)

	var entities []uid.UID64
	factory.Create(count)
	for factory.Next() {
		entities = append(entities, factory.IDs...)
	}
	return entities
}

// Benchmark_Editor_Del exercises a structural remove through the
// production CmdBuf path: entities carry 10 data components and each leaf
// removes comps=N of them; only the Sync executing Editor.Migrate is
// timed. The comps=10 world anchors entities with an extra Base component
// (via populateWithBase) so removing all 10 tracked components never
// unlinks them. subset<pop leaves additionally exercise source compaction.
func Benchmark_Editor_Del(b *testing.B) {
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
				fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=%d", entitiesNumber, subset, n),
				subset,
				func() (*goke.Query, *goke.Editor, goke.Runnable) {
					var migrateQ, restoreQ *goke.Query
					var fwd, rev *goke.Editor
					ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
						if n == 10 {
							populateWithBase(si, entitiesNumber)
						} else {
							populate(si, entitiesNumber)
						}
						migrateQ = si.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
						if n == 10 {
							restoreQ = si.NewQueryBuilder().Include(goke.Include[Base]()).Exclude(goke.Exclude[Pos]()).Build()
						} else {
							restoreQ = si.NewQueryBuilder().Include(goke.Include[T10]()).Exclude(goke.Exclude[Pos]()).Build()
						}
						fwd = migrateQ.NewEditorBuilder().Remove(delComps[:n]...).Build()
						rev = restoreQ.NewEditorBuilder(addComps[:n]...).Build()
					}})
					return migrateQ, fwd, ecs.RegSys(enqueueAll(restoreQ, rev))
				})
		}
	}
}
