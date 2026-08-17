package bench_test

import (
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

const entitiesNumber = 1024
const filterSubsetSize = 512

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

// populateBase spawns count Base-only entities. Must be called from within
// the caller's own Setup (si comes from that Setup's OnInit) — it does not
// call Setup itself, since Setup is callable only once per ECS.
func populateBase(si *goke.SysInit, count int) []uid.UID64 {
	factory := si.NewFactory(new(goke.Comp[Base]))
	var ids []uid.UID64
	factory.Create(count)
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}
	return ids
}

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

// populator holds the Factory and component handles for spawning the
// standard 10-data-component entity shape. Build once via newPopulator
// (SysInit-gated), then call spawn as many times as needed — spawn itself
// only uses the already-built Factory, so it carries no Setup restriction.
type populator struct {
	factory *goke.Factory
	c1      goke.Comp[Pos]
	c2      goke.Comp[Vel]
	c3      goke.Comp[Acc]
	c4      goke.Comp[T04]
	c5      goke.Comp[T05]
	c6      goke.Comp[T06]
	c7      goke.Comp[T07]
	c8      goke.Comp[T08]
	c9      goke.Comp[T09]
	c10     goke.Comp[T10]
}

// newPopulator builds a populator. Must be called from within the caller's
// own Setup — see populateBase.
func newPopulator(si *goke.SysInit) *populator {
	p := &populator{}
	p.factory = si.NewFactory(&p.c1, &p.c2, &p.c3, &p.c4, &p.c5, &p.c6, &p.c7, &p.c8, &p.c9, &p.c10)
	return p
}

// spawn creates count entities with all 10 data components, randomized.
// Safe to call repeatedly, including outside any Setup.
func (p *populator) spawn(count int) []uid.UID64 {
	var entities []uid.UID64
	p.factory.Create(count)
	for p.factory.Next() {
		comp1 := p.c1.Slice(&p.factory.Cursor)
		comp2 := p.c2.Slice(&p.factory.Cursor)
		comp3 := p.c3.Slice(&p.factory.Cursor)
		comp4 := p.c4.Slice(&p.factory.Cursor)
		comp5 := p.c5.Slice(&p.factory.Cursor)
		comp6 := p.c6.Slice(&p.factory.Cursor)
		comp7 := p.c7.Slice(&p.factory.Cursor)
		comp8 := p.c8.Slice(&p.factory.Cursor)
		comp9 := p.c9.Slice(&p.factory.Cursor)
		comp10 := p.c10.Slice(&p.factory.Cursor)
		for i, entityID := range p.factory.IDs {
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

// populate spawns count entities with all 10 data components, randomized —
// a one-shot convenience over populator for callers that never need to
// spawn again. Must be called from within the caller's own Setup.
func populate(si *goke.SysInit, count int) []uid.UID64 {
	return newPopulator(si).spawn(count)
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

// benchOrders are the two entity-order variants structural benchmarks run in:
// sorted iterates entities in creation order, random reshuffles them before
// every iteration (outside the timer).
var benchOrders = []string{"sorted", "random"}

func shuffleIDs(r *rand.Rand, ids []uid.UID64) {
	r.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
}
