package bench_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke/v3"
)

type HitTag struct{ V float32 }

// Benchmark_Optional_vs_TwoQueries compares the two ways to read a component
// that only some entities carry: two disjoint Include/Exclude queries versus
// one query tracking it via Optional. Population is split 50/50 between a
// Pos-only archetype and a Pos+HitTag archetype. Follows Benchmark_Matcher_All's
// write-in-place / hoisted-slice rules to keep the numbers meaningful.
func Benchmark_Optional_vs_TwoQueries(b *testing.B) {
	ecs := setupECS()
	_ = ecs.RegComp[HitTag]()

	var queryPlain, queryHit, queryOpt *goke.Query
	var posPlain, posHit, posOpt goke.Comp[Pos]
	var hitHit goke.Comp[HitTag]
	var hitOpt goke.OptComp[HitTag]

	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		half := entitiesNumber / 2

		var spawnPos goke.Comp[Pos]
		plainFactory := si.NewFactory(&spawnPos)
		plainFactory.Create(half)
		for plainFactory.Next() {
			ps := spawnPos.Slice(&plainFactory.Cursor)
			for i := range plainFactory.IDs {
				ps[i] = Pos{rand.Float32() * 100, rand.Float32() * 100}
			}
		}

		var spawnPosHit goke.Comp[Pos]
		var spawnHit goke.Comp[HitTag]
		hitFactory := si.NewFactory(&spawnPosHit, &spawnHit)
		hitFactory.Create(half)
		for hitFactory.Next() {
			ps := spawnPosHit.Slice(&hitFactory.Cursor)
			hs := spawnHit.Slice(&hitFactory.Cursor)
			for i := range hitFactory.IDs {
				ps[i] = Pos{rand.Float32() * 100, rand.Float32() * 100}
				hs[i] = HitTag{V: 1}
			}
		}

		queryPlain = si.NewQueryBuilder(&posPlain).Exclude(goke.Exclude[HitTag]()).Build()
		queryHit = si.NewQueryBuilder(&posHit, &hitHit).Build()
		queryOpt = si.NewQueryBuilder(&posOpt).Optional(&hitOpt).Build()
	}})

	b.Run(fmt.Sprintf("pop=%d/TwoQueries", entitiesNumber), func(b *testing.B) {
		fn := func() {
			queryPlain.All()
			for queryPlain.Next() {
				ps := posPlain.Slice(queryPlain.Cursor())
				for i := range queryPlain.Cursor().IDs {
					ps[i].X += ps[i].Y
				}
			}
			queryHit.All()
			for queryHit.Next() {
				ps := posHit.Slice(queryHit.Cursor())
				hs := hitHit.Slice(queryHit.Cursor())
				for i := range queryHit.Cursor().IDs {
					ps[i].X += ps[i].Y
					hs[i].V += 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/OptionalQuery", entitiesNumber), func(b *testing.B) {
		fn := func() {
			queryOpt.All()
			for queryOpt.Next() {
				cur := queryOpt.Cursor()
				ps := posOpt.Slice(cur)
				if hitOpt.Present(cur) {
					hs := hitOpt.Slice(cur)
					for i := range cur.IDs {
						ps[i].X += ps[i].Y
						hs[i].V += 1
					}
				} else {
					for i := range cur.IDs {
						ps[i].X += ps[i].Y
					}
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
