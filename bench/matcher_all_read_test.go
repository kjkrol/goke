package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke/v3"
)

// Benchmark_Matcher_All_Read compares reading two of three columns through Comp (rw) and
// CompRO (read), the same loop otherwise; see Benchmark_Matcher_All for the rules it keeps.
func Benchmark_Matcher_All_Read(b *testing.B) {
	ecs := setupECS()

	var pos goke.Comp[Pos]
	var vel goke.Comp[Vel]
	var acc goke.Comp[Acc]
	var velR goke.CompRO[Vel]
	var accR goke.CompRO[Acc]

	var rw, read *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		populate(si, entitiesNumber)
		rw = si.NewQueryBuilder(&pos, &vel, &acc).Build()
		read = si.NewQueryBuilder(&pos, &velR, &accR).Build()
	}})

	b.Run(fmt.Sprintf("pop=%d/rw", entitiesNumber), func(b *testing.B) {
		cursor := rw.Cursor()
		fn := func() {
			rw.All()
			for rw.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				for i, entityID := range cursor.IDs {
					_ = entityID
					posSlice[i].X += velSlice[i].X + accSlice[i].X
					posSlice[i].Y += velSlice[i].Y + accSlice[i].Y
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/read", entitiesNumber), func(b *testing.B) {
		cursor := read.Cursor()
		fn := func() {
			read.All()
			for read.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := velR.Slice(cursor)
				accSlice := accR.Slice(cursor)
				for i, entityID := range cursor.IDs {
					_ = entityID
					v, a := velSlice.At(i), accSlice.At(i)
					posSlice[i].X += v.X + a.X
					posSlice[i].Y += v.Y + a.Y
				}
			}
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})
}
