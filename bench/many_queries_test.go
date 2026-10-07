package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v3"
)

// manyQueries are how many other queries the ECS holds while one is measured: a game's plugins
// register dozens. The catalog holds at most 64 matchers before growable matchers.
var manyQueries = []int{8, 56}

// registerQueries builds n queries of rotating shapes over the standard components, none of them
// iterated: the catalog holds a matcher for each.
func registerQueries(si *goke.SysInit, n int) {
	var pos goke.Comp[Pos]
	var vel goke.Comp[Vel]
	var acc goke.Comp[Acc]
	var t04 goke.Comp[T04]
	var t05 goke.Comp[T05]
	var t06 goke.Comp[T06]
	shapes := [][]goke.Trackable{
		{&pos}, {&pos, &vel}, {&vel, &acc}, {&acc, &t04}, {&t04, &t05, &t06}, {&pos, &t05}, {&vel, &t06},
	}
	for i := range n {
		si.NewQueryBuilder(shapes[i%len(shapes)]...).Build()
	}
}

// Benchmark_ManyQueries_All measures iterating one query of two components over the population
// while n other queries are registered on the same ECS — a query walks its own matcher, so n
// should not show; it guards how the catalog keeps matchers in memory.
func Benchmark_ManyQueries_All(b *testing.B) {
	for _, n := range manyQueries {
		ecs := setupECS()
		var pos goke.Comp[Pos]
		var vel goke.Comp[Vel]
		var query *goke.Query
		ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
			registerQueries(si, n)
			populate(si, entitiesNumber)
			query = si.NewQueryBuilder(&pos, &vel).Build()
		}})

		var count int
		b.Run(fmt.Sprintf("pop=%d/queries=%d", entitiesNumber, n), func(b *testing.B) {
			cursor := query.Cursor()
			fn := func() {
				count = 0
				query.All()
				for query.Next() {
					posSlice := pos.Slice(cursor)
					velSlice := vel.Slice(cursor)
					for i := range cursor.IDs {
						posSlice[i].X += velSlice[i].X
						count++
					}
				}
			}
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					fn()
				}
			})
			if count != entitiesNumber {
				b.Fatalf("sanity check failed: expected %d, got %d", entitiesNumber, count)
			}
		})
	}
}
