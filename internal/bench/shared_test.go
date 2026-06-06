package bench_test

import (
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/internal/core"
)

const entitiesNumber = 1000

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
	_ = goke.RegisterComponent[Pos](ecs)
	_ = goke.RegisterComponent[Vel](ecs)
	_ = goke.RegisterComponent[Acc](ecs)
	_ = goke.RegisterComponent[T04](ecs)
	_ = goke.RegisterComponent[T05](ecs)
	_ = goke.RegisterComponent[T06](ecs)
	_ = goke.RegisterComponent[T07](ecs)
	_ = goke.RegisterComponent[T08](ecs)
	_ = goke.RegisterComponent[T09](ecs)
	_ = goke.RegisterComponent[T10](ecs)
	_ = goke.RegisterComponent[Tag](ecs)
	return ecs
}

func populate(ecs *goke.ECS, count int) []core.Entity {
	var entities []core.Entity
	blueprint := goke.NewBlueprint10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
	for page := range blueprint.Create(count) {
		for i, entity := range page.Entity {
			page.Comp1[i] = Pos{rand.Float32() * 100, rand.Float32() * 100}
			page.Comp2[i] = Vel{rand.Float32() * 40, 1}
			page.Comp3[i] = Acc{rand.Float32(), 0.1}
			page.Comp4[i] = T04{rand.Float32()}
			page.Comp5[i] = T05{rand.Float32()}
			page.Comp6[i] = T06{rand.Float32()}
			page.Comp7[i] = T07{rand.Float64()}
			page.Comp8[i] = T08{rand.Float64()}
			page.Comp9[i] = T09{rand.Float64()}
			page.Comp10[i] = T10{rand.Float64()}

			entities = append(entities, entity)
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
