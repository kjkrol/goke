package bench

import (
	"testing"

	"github.com/kjkrol/goke"
)

type Position3 struct{ X, Y, Z float64 }
type Velocity3 struct{ X, Y, Z float64 }
type ProcessedTag struct{}
type Props1 struct{ V float64 }
type Props2 struct{ V float64 }
type Props3 struct{ V float64 }
type Props4 struct{ V float64 }

func BenchmarkEngine_Structural(b *testing.B) {
	ecs := goke.New(
		goke.WithInitialEntityCap(100000),
		goke.WithFreeIndicesCap(100000),
	)
	// Pre-registering for "Turbo" performance in bench

	// --- 1. ENTITY CREATION & INITIALIZATION ---
	// Tests CreateEntity
	b.Run("Create Entity With 1 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint1[Props1](ecs)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, a := blueprint.Create()
			a.V = 1
		}
	})
	b.Run("Create Entity With 2 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint2[Props1, Props2](ecs)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, a, b := blueprint.Create()
			a.V = 1
			b.V = 2
		}
	})
	b.Run("Create Entity With 3 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint3[Props1, Props2, Props3](ecs)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, a, b, c := blueprint.Create()
			a.V = 1
			b.V = 2
			c.V = 3
		}
	})
	b.Run("Create Entity With 4 comp", func(b *testing.B) {
		goke.Reset(ecs)
		blueprint := goke.NewBlueprint4[Props1, Props2, Props3, Props4](ecs)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, a, b, c, d := blueprint.Create()
			a.V = 1
			b.V = 2
			c.V = 3
			d.V = 4
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Velocity3} to {Velocity3, Position3}
	b.Run("Add 2nd Component", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]goke.Entity, b.N)

		posDesc := goke.RegisterComponent[Position3](ecs)
		blueprintA := goke.NewBlueprint1[Velocity3](ecs)
		_, _ = blueprintA.Create()

		for i := 0; i < b.N; i++ {
			entities[i], _ = blueprintA.Create()
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			*goke.EnsureComponent[Position3](ecs, entities[i], posDesc) = Position3{X: 10, Y: 10}
		}
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add Tag", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]goke.Entity, b.N)

		tagDesc := goke.RegisterComponent[ProcessedTag](ecs)
		blueprintA := goke.NewBlueprint1[Velocity3](ecs)
		_, _ = blueprintA.Create()

		for i := 0; i < b.N; i++ {
			entities[i], _ = blueprintA.Create()
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			goke.EnsureComponent[ProcessedTag](ecs, entities[i], tagDesc)
		}
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove Component", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]goke.Entity, b.N)

		posDesc := goke.RegisterComponent[Position3](ecs)
		blueprintA := goke.NewBlueprint1[Velocity3](ecs)
		_, _ = blueprintA.Create()

		for i := 0; i < b.N; i++ {
			entities[i], _ = blueprintA.Create()
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

		velDesc := goke.RegisterComponent[Velocity3](ecs)
		blueprintA := goke.NewBlueprint1[Velocity3](ecs)
		_, _ = blueprintA.Create()

		e, vel := blueprintA.Create()
		*vel = Velocity3{X: 1, Y: 2, Z: 3}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pos := goke.GetComponent[Velocity3](ecs, e, velDesc)
			pos.X += 1.0
		}
	})

	// Benchmark for ComponentGet: Uses reflection to find ComponentInfo.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component (Safe)", func(b *testing.B) {
		goke.Reset(ecs)

		velDesc := goke.RegisterComponent[Velocity3](ecs)
		blueprintA := goke.NewBlueprint1[Velocity3](ecs)
		_, _ = blueprintA.Create()

		e, vel := blueprintA.Create()
		*vel = Velocity3{X: 1, Y: 2, Z: 3}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pos, err := goke.SafeGetComponent[Velocity3](ecs, e, velDesc)
			if err == nil {
				pos.X += 1.0
			}
		}
	})

	// Benchmark for ComponentGet: Uses reflection to find ComponentInfo.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component via View.Filter", func(b *testing.B) {
		goke.Reset(ecs)

		_ = goke.RegisterComponent[Velocity3](ecs)
		blueprintA := goke.NewBlueprint1[Velocity3](ecs)
		_, _ = blueprintA.Create()

		e, vel := blueprintA.Create()
		*vel = Velocity3{X: 1, Y: 2, Z: 3}

		view := goke.NewView1[Velocity3](ecs)
		arr := []goke.Entity{e}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for head := range view.FilterValues(arr) {
				pos := head.V1
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
	_ = goke.RegisterComponent[Position3](ecs)
	blueprint := goke.NewBlueprint1[Position3](ecs)
	count := 100000
	for i := 0; i < count; i++ {
		_, pos := blueprint.Create()
		*pos = Position3{X: float64(i)}
	}

	view := goke.NewView1[Position3](ecs)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for head := range view.All() {
			head.V1.X += 0.1
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
	_ = goke.RegisterComponent[Position3](ecs)

	// --- SETUP PHASE ---
	// Pre-create entities outside the timed loop to isolate the cost of removal
	blueprint := goke.NewBlueprint1[Position3](ecs)
	entities := make([]goke.Entity, count)
	for j := 0; j < count; j++ {
		entities[j], _ = blueprint.Create()
	}

	b.ResetTimer() // Exclude setup time from the results
	b.ReportAllocs()

	// --- EXECUTION PHASE ---
	for i := 0; i < b.N; i++ {
		idx := i % count

		// If b.N > count, we need to re-create the entity to ensure
		// we are benchmarking a real 'Remove' operation rather than
		// an early-exit check for a non-existent entity.
		if i >= count && i%count == 0 {
			b.StopTimer()
			for j := 0; j < count; j++ {
				entities[j], _ = blueprint.Create()
			}
			b.StartTimer()
		}

		goke.RemoveEntity(ecs, entities[idx])
	}
}

func BenchmarkEngine_AddRemove_Stability(b *testing.B) {
	ecs := goke.New(goke.WithInitialEntityCap(1024))
	_ = goke.RegisterComponent[Position3](ecs)
	blueprint := goke.NewBlueprint1[Position3](ecs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, pos := blueprint.Create()
		*pos = Position3{X: 1}

		// Usuwamy co drugą, żeby wymusić swapowanie w archetypie
		if i%2 == 0 {
			goke.RemoveEntity(ecs, e)
		}
	}
}
