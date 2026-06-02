package bench

import (
	"testing"

	"github.com/kjkrol/goke"
)

// --- Benchmark CreateBatch ---

const benchSize = 1000

func BenchmarkBlueprint1_CreateBatch(b *testing.B) {
	b.StopTimer()
	ecs := goke.New()
	_ = goke.RegisterComponent[Pos](ecs)

	blueprint := goke.NewBlueprint1[Pos](ecs)
	buf := make([]goke.Item1[Pos], benchSize)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		count := 0
		for _, item := range blueprint.CreateBatch(entitiesNumber, buf) {
			// Prosta operacja write/read, aby zapobiec dead-code elimination przez kompilator
			pos := item.V1
			pos.X += 1.0
			count++
		}
		GlobalCount = count
	}

	if GlobalCount != entitiesNumber {
		b.Fatalf("Blueprint1 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
	}
}

func BenchmarkBlueprint10_CreateBatch(b *testing.B) {
	b.StopTimer()
	ecs := goke.New()
	_ = goke.RegisterComponent[Pos](ecs)
	_ = goke.RegisterComponent[Vel](ecs)
	_ = goke.RegisterComponent[Acc](ecs)
	_ = goke.RegisterComponent[Mass](ecs)
	_ = goke.RegisterComponent[Spin](ecs)
	_ = goke.RegisterComponent[Char](ecs)
	_ = goke.RegisterComponent[Elec](ecs)
	_ = goke.RegisterComponent[Magn](ecs)
	_ = goke.RegisterComponent[T09](ecs)
	_ = goke.RegisterComponent[T10](ecs)

	blueprint := goke.NewBlueprint10[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn, T09, T10](ecs)
	buf := make([]goke.Item10[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn, T09, T10], benchSize)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		count := 0
		for _, item := range blueprint.CreateBatch(entitiesNumber, buf) {
			// Dostęp do pierwszego i ostatniego komponentu w celu wymuszenia ewaluacji wskaźników
			pos := item.V1
			pos.X += 1.0

			v10 := item.V10
			v10.V += 1.0

			count++
		}
		GlobalCount = count
	}

	if GlobalCount != entitiesNumber {
		b.Fatalf("Blueprint10 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
	}
}
