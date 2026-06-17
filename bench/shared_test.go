package bench_test

import (
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/kjkrol/goke"
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

func setupECS() *goke.ECS {
	ecs := goke.New()
	_ = goke.RegCompType[Pos](ecs)
	_ = goke.RegCompType[Vel](ecs)
	_ = goke.RegCompType[Acc](ecs)
	_ = goke.RegCompType[T04](ecs)
	_ = goke.RegCompType[T05](ecs)
	_ = goke.RegCompType[T06](ecs)
	_ = goke.RegCompType[T07](ecs)
	_ = goke.RegCompType[T08](ecs)
	_ = goke.RegCompType[T09](ecs)
	_ = goke.RegCompType[T10](ecs)
	_ = goke.RegCompType[Tag](ecs)
	return ecs
}

func populate(ecs *goke.ECS, count int) []uid.UID64 {
	var c1 goke.Col[Pos]
	var c2 goke.Col[Vel]
	var c3 goke.Col[Acc]
	var c4 goke.Col[T04]
	var c5 goke.Col[T05]
	var c6 goke.Col[T06]
	var c7 goke.Col[T07]
	var c8 goke.Col[T08]
	var c9 goke.Col[T09]
	var c10 goke.Col[T10]
	blueprint := goke.CreateEntFactory(ecs, goke.Track(&c1), goke.Track(&c2), goke.Track(&c3), goke.Track(&c4), goke.Track(&c5), goke.Track(&c6), goke.Track(&c7), goke.Track(&c8), goke.Track(&c9), goke.Track(&c10))

	var entities []uid.UID64
	blueprint.Create(count)
	for blueprint.Next() {
		comp1 := c1.Slice(&blueprint.Cursor)
		comp2 := c2.Slice(&blueprint.Cursor)
		comp3 := c3.Slice(&blueprint.Cursor)
		comp4 := c4.Slice(&blueprint.Cursor)
		comp5 := c5.Slice(&blueprint.Cursor)
		comp6 := c6.Slice(&blueprint.Cursor)
		comp7 := c7.Slice(&blueprint.Cursor)
		comp8 := c8.Slice(&blueprint.Cursor)
		comp9 := c9.Slice(&blueprint.Cursor)
		comp10 := c10.Slice(&blueprint.Cursor)
		for i, entityID := range blueprint.IDs {
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
