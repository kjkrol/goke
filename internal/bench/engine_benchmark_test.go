package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
)

func BenchmarkEngine_Structural(b *testing.B) {
	ecs := goke.New()

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	b.Run("Add 2nd Component", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]goke.Entity, b.N)

		posDesc := goke.RegisterComponent[Pos](ecs)
		blueprintA := goke.NewBlueprint1[Vel](ecs)
		blueprintA.Create()

		for i := 0; i < b.N; i++ {
			entities[i] = blueprintA.Create().Entity
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			*goke.EnsureComponent[Pos](ecs, entities[i], posDesc) = Pos{X: 10, Y: 10}
		}
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add Tag", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]goke.Entity, b.N)

		tagDesc := goke.RegisterComponent[Tag](ecs)
		blueprintA := goke.NewBlueprint1[Vel](ecs)
		blueprintA.Create()

		for i := 0; i < b.N; i++ {
			entities[i] = blueprintA.Create().Entity
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			goke.EnsureComponent[Tag](ecs, entities[i], tagDesc)
		}
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove Component", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]goke.Entity, b.N)

		posDesc := goke.RegisterComponent[Pos](ecs)
		blueprintA := goke.NewBlueprint1[Vel](ecs)
		blueprintA.Create()

		for i := 0; i < b.N; i++ {
			entities[i] = blueprintA.Create().Entity
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			goke.RemoveComponent(ecs, entities[i], posDesc)
		}
	})

	// --- 5. COMPONENT ACCESS (GET) ---

	// Benchmark for ComponentGet: Uses reflection to find ComponentInfo.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component", func(b *testing.B) {
		goke.Reset(ecs)

		velDesc := goke.RegisterComponent[Vel](ecs)
		blueprintA := goke.NewBlueprint1[Vel](ecs)
		blueprintA.Create()

		item := blueprintA.Create()
		e, vel := item.Entity, item.Comp1
		*vel = Vel{X: 1, Y: 2}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pos := goke.GetComponent[Vel](ecs, e, velDesc)
			pos.X += 1.0
		}
	})

	// Benchmark for ComponentGet: Uses reflection to find ComponentInfo.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component (Safe)", func(b *testing.B) {
		goke.Reset(ecs)

		velDesc := goke.RegisterComponent[Vel](ecs)
		blueprintA := goke.NewBlueprint1[Vel](ecs)
		blueprintA.Create()

		item := blueprintA.Create()
		e, vel := item.Entity, item.Comp1
		*vel = Vel{X: 1, Y: 2}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pos, err := goke.SafeGetComponent[Vel](ecs, e, velDesc)
			if err == nil {
				pos.X += 1.0
			}
		}
	})

	// Benchmark for ComponentGet: Uses reflection to find ComponentInfo.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component via View.Filter", func(b *testing.B) {
		goke.Reset(ecs)

		_ = goke.RegisterComponent[Vel](ecs)
		blueprintA := goke.NewBlueprint1[Vel](ecs)
		blueprintA.Create()

		item := blueprintA.Create()
		e, vel := item.Entity, item.Comp1
		*vel = Vel{X: 1, Y: 2}

		view := goke.NewView1[Vel](ecs)
		arr := []goke.Entity{e}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for item := range view.Filter(arr) {
				pos := item.Comp1
				pos.X += 1.0
			}
		}
	})
}

func BenchmarkEngine_Query(b *testing.B) {
	ecs := goke.New(
		goke.WithInitialEntityCap(100000),
		goke.WithFreeIndicesCap(100000),
	)
	_ = goke.RegisterComponent[Pos](ecs)
	blueprint := goke.NewBlueprint1[Pos](ecs)
	count := 100000
	for i := range count {
		item := blueprint.Create()
		_, pos := item.Entity, item.Comp1
		*pos = Pos{X: float32(i)}
	}

	view := goke.NewView1[Pos](ecs)

	b.ReportAllocs()
	for b.Loop() {
		for item := range view.All() {
			pos := item.Comp1
			pos.X += 0.1
		}
	}
}

func BenchmarkEngine_RemoveEntity_Clean(b *testing.B) {
	count := 100000
	// Initialize the ecs with "Turbo" settings to pre-allocate memory buffers
	ecs := goke.New(
		goke.WithInitialEntityCap(count),
		goke.WithFreeIndicesCap(count),
	)
	_ = goke.RegisterComponent[Pos](ecs)

	// --- SETUP PHASE ---
	// Pre-create entities outside the timed loop to isolate the cost of removal
	blueprint := goke.NewBlueprint1[Pos](ecs)
	entities := make([]goke.Entity, count)
	for j := range count {
		entities[j] = blueprint.Create().Entity
	}

	// Exclude setup time from the results
	b.ReportAllocs()

	// --- EXECUTION PHASE ---
	for i := 0; b.Loop(); i++ {
		idx := i % count

		// If b.N > count, we need to re-create the entity to ensure
		// we are benchmarking a real 'Remove' operation rather than
		// an early-exit check for a non-existent entity.
		if i >= count && i%count == 0 {
			b.StopTimer()
			for j := range count {
				entities[j] = blueprint.Create().Entity
			}
			b.StartTimer()
		}

		goke.RemoveEntity(ecs, entities[idx])
	}
}

func BenchmarkEngine_AddRemove_Stability(b *testing.B) {
	ecs := goke.New(goke.WithInitialEntityCap(1024))
	_ = goke.RegisterComponent[Pos](ecs)
	blueprint := goke.NewBlueprint1[Pos](ecs)

	for i := 0; b.Loop(); i++ {
		item := blueprint.Create()
		e, pos := item.Entity, item.Comp1
		*pos = Pos{X: 1}

		// Usuwamy co drugą, żeby wymusić swapowanie w archetypie
		if i%2 == 0 {
			goke.RemoveEntity(ecs, e)
		}
	}
}
