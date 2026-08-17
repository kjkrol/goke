package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// E01..E10 are the incoming component types used by Benchmark_Editor_Mix.
// For N=K the migrator adds E01..EK and removes the K existing components,
// replacing them completely in a single archetype migration.
type E01 struct{ V float32 }
type E02 struct{ V float32 }
type E03 struct{ V float32 }
type E04 struct{ V float32 }
type E05 struct{ V float32 }
type E06 struct{ V float32 }
type E07 struct{ V float32 }
type E08 struct{ V float32 }
type E09 struct{ V float32 }
type E10 struct{ V float32 }

func setupEditECS() *goke.ECS {
	ecs := setupECS()
	_ = goke.RegComp[E01](ecs)
	_ = goke.RegComp[E02](ecs)
	_ = goke.RegComp[E03](ecs)
	_ = goke.RegComp[E04](ecs)
	_ = goke.RegComp[E05](ecs)
	_ = goke.RegComp[E06](ecs)
	_ = goke.RegComp[E07](ecs)
	_ = goke.RegComp[E08](ecs)
	_ = goke.RegComp[E09](ecs)
	_ = goke.RegComp[E10](ecs)
	return ecs
}

func populateN(si *goke.SysInit, count, n int) []uid.UID64 {
	var ids []uid.UID64
	switch n {
	case 1:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 2:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 3:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 4:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 5:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 6:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 7:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 8:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07]), new(goke.Comp[T08]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 9:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07]), new(goke.Comp[T08]), new(goke.Comp[T09]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 10:
		f := si.NewFactory(new(goke.Comp[Base]), new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07]), new(goke.Comp[T08]), new(goke.Comp[T09]), new(goke.Comp[T10]))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	}
	return ids
}

// Benchmark_Editor_Mix exercises a combined add+remove through the
// production CmdBuf path: each leaf swap=N adds N new types (E01..E0N) and
// removes N existing ones (Pos..) in a single bulk migration; only the Sync
// executing Editor.Migrate is timed. subset<pop leaves additionally
// exercise source compaction.
func Benchmark_Editor_Mix(b *testing.B) {
	ecs := setupEditECS()
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
	addE := []goke.Addable{
		new(goke.Comp[E01]), new(goke.Comp[E02]), new(goke.Comp[E03]),
		new(goke.Comp[E04]), new(goke.Comp[E05]), new(goke.Comp[E06]),
		new(goke.Comp[E07]), new(goke.Comp[E08]), new(goke.Comp[E09]),
		new(goke.Comp[E10]),
	}
	delE := []goke.EditOpt{
		goke.Remove[E01](), goke.Remove[E02](), goke.Remove[E03](),
		goke.Remove[E04](), goke.Remove[E05](), goke.Remove[E06](),
		goke.Remove[E07](), goke.Remove[E08](), goke.Remove[E09](),
		goke.Remove[E10](),
	}

	for n := 1; n <= 10; n++ {
		for _, subset := range []int{filterSubsetSize, entitiesNumber} {
			runEditorLeaf(b, ecs,
				fmt.Sprintf("pop=%d/subset=%d/comp=%d/swap=%d", entitiesNumber, subset, n+1, n),
				subset,
				func() (*goke.Query, *goke.Editor, goke.Runnable) {
					var migrateQ, restoreQ *goke.Query
					var fwd, rev *goke.Editor
					ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
						populateN(si, entitiesNumber, n)
						migrateQ = si.NewQueryBuilder().Include(goke.Include[Pos]()).Build()
						restoreQ = si.NewQueryBuilder().Include(goke.Include[E01]()).Build()
						fwd = migrateQ.NewEditorBuilder(addE[:n]...).Remove(delComps[:n]...).Build()
						rev = restoreQ.NewEditorBuilder(addComps[:n]...).Remove(delE[:n]...).Build()
					}})
					return migrateQ, fwd, ecs.RegSys(enqueueAll(restoreQ, rev))
				})
		}
	}
}
