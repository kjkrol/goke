package bench_test

import (
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

const entitiesNumber = 1024
const filterSubsetSize = 100

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }
type T04 struct{ V float32 }
type T05 struct{ V float32 }
type T06 struct{ V float32 }
type T07 struct{ V float64 }
type T08 struct{ V float64 }
type T09 struct{ V float64 }
type T10 struct{ V float64 }

type Tag struct{}

type Base struct{ V int32 }

func setupECS() *goke.ECS {
	ecs := goke.New()
	_ = goke.RegComp[Pos](ecs)
	_ = goke.RegComp[Vel](ecs)
	_ = goke.RegComp[Acc](ecs)
	_ = goke.RegComp[T04](ecs)
	_ = goke.RegComp[T05](ecs)
	_ = goke.RegComp[T06](ecs)
	_ = goke.RegComp[T07](ecs)
	_ = goke.RegComp[T08](ecs)
	_ = goke.RegComp[T09](ecs)
	_ = goke.RegComp[T10](ecs)
	_ = goke.RegComp[Tag](ecs)
	_ = goke.RegComp[Base](ecs)
	return ecs
}

func populate(ecs *goke.ECS, count int) []uid.UID64 {
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
	factory := ecs.NewFactory(&c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10)

	var entities []uid.UID64
	factory.Create(count)
	for factory.Next() {
		comp1 := c1.Slice(&factory.Cursor)
		comp2 := c2.Slice(&factory.Cursor)
		comp3 := c3.Slice(&factory.Cursor)
		comp4 := c4.Slice(&factory.Cursor)
		comp5 := c5.Slice(&factory.Cursor)
		comp6 := c6.Slice(&factory.Cursor)
		comp7 := c7.Slice(&factory.Cursor)
		comp8 := c8.Slice(&factory.Cursor)
		comp9 := c9.Slice(&factory.Cursor)
		comp10 := c10.Slice(&factory.Cursor)
		for i, entityID := range factory.IDs {
			comp1[i] = Pos{rand.Float32() * 100, rand.Float32() * 100}
			comp2[i] = Vel{rand.Float32() * 40, 1}
			comp3[i] = Acc{rand.Float32(), 0.1}
			comp4[i] = T04{rand.Float32()}
			comp5[i] = T05{rand.Float32()}
			comp6[i] = T06{rand.Float32()}
			comp7[i] = T07{rand.Float64()}
			comp8[i] = T08{rand.Float64()}
			comp9[i] = T09{rand.Float64()}
			comp10[i] = T10{rand.Float64()}
			entities = append(entities, entityID)
		}
	}
	return entities
}

func measurePerEntity(b *testing.B, batchSize int, benchLoop func()) {
	var mStart, mEnd runtime.MemStats

	// Wymuszamy odśmiecacz pamięci, żeby usunąć pozostałości po fazie setupu
	runtime.GC()
	runtime.ReadMemStats(&mStart)

	b.ResetTimer()

	// Wykonujemy właściwą pętlę b.N (0 narzutu na każdą iterację, bo wywołujemy to raz)
	benchLoop()

	b.StopTimer()
	runtime.ReadMemStats(&mEnd)

	// Obliczenia
	totalEntities := float64(b.N * batchSize)
	elapsedNs := float64(b.Elapsed().Nanoseconds())
	// elapsedSec := b.Elapsed().Seconds()

	// allocBytes := float64(mEnd.TotalAlloc - mStart.TotalAlloc)
	// allocs := float64(mEnd.Mallocs - mStart.Mallocs)

	// Raportowanie
	b.ReportMetric(elapsedNs/totalEntities, "ns/ent")
	// b.ReportMetric(totalEntities/elapsedSec, "ent/s")
	// b.ReportMetric(allocBytes/totalEntities, "B/ent")
	// b.ReportMetric(allocs/totalEntities, "allocs/ent")
}
