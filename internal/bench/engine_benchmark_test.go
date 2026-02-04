package bench

import (
	"testing"

	"github.com/kjkrol/goke"
)

type Position3 struct{ X, Y, Z float64 }
type Velocity3 struct{ X, Y, Z float64 }
type ProcessedTag struct{}

func BenchmarkEngine_Structural(b *testing.B) {
	ecs := goke.New(
		goke.WithInitialEntityCap(100000),
		goke.WithDefaultArchetypeChunkSize(4096),
		goke.WithInitialArchetypeRegistryCap(128),
		goke.WithFreeIndicesCap(100000),
		goke.WithViewRegistryInitCap(64),
	)

	// Pre-registering for "Turbo" performance in bench
	posDesc := goke.RegisterComponent[Position3](ecs)
	velDesc := goke.RegisterComponent[Velocity3](ecs)
	tagDesc := goke.RegisterComponent[ProcessedTag](ecs)

	// initialize archetypes
	eStart1 := goke.CreateEntity(ecs)
	goke.EnsureComponent[Velocity3](ecs, eStart1, velDesc)

	eStart2 := goke.CreateEntity(ecs)
	goke.EnsureComponent[Position3](ecs, eStart2, posDesc)
	goke.EnsureComponent[Velocity3](ecs, eStart2, velDesc)

	eStart3 := goke.CreateEntity(ecs)
	goke.EnsureComponent[ProcessedTag](ecs, eStart3, tagDesc)

	// --- 1. ENTITY CREATION & INITIALIZATION ---
	// Tests CreateEntity
	b.Run("Create_Entity", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			e := goke.CreateEntity(ecs)
			_ = e
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Position} to {Position, Velocity}
	b.Run("Add_Component", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.CreateEntity(ecs)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			*goke.EnsureComponent[Velocity3](ecs, entities[i], velDesc) = Velocity3{X: 10, Y: 10}
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Position} to {Position, Velocity}
	b.Run("Add_2nd_Component", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.CreateEntity(ecs)
			*goke.EnsureComponent[Position3](ecs, entities[i], posDesc) = Position3{X: 1}
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			*goke.EnsureComponent[Velocity3](ecs, entities[i], velDesc) = Velocity3{X: 10, Y: 10}
		}
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add_Tag", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.CreateEntity(ecs)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			goke.EnsureComponent[ProcessedTag](ecs, entities[i], tagDesc)
		}
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove_Component", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.CreateEntity(ecs)
			goke.EnsureComponent[Position3](ecs, entities[i], posDesc)
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
	b.Run("Get_Component_Reflect", func(b *testing.B) {
		e := goke.CreateEntity(ecs)
		*goke.EnsureComponent[Position3](ecs, e, posDesc) = Position3{X: 1, Y: 2, Z: 3}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Internally calls reflect.TypeFor[T]() and registry lookups
			pos, err := goke.GetComponent[Position3](ecs, e, posDesc)
			if err == nil {
				pos.X += 1.0
			}
		}
	})

	// Benchmark for GetComponentByType: Uses pre-fetched ComponentInfo.
	// This is the "Fast Path" - bypasses reflection and registry maps.
	b.Run("Get_Component_Direct", func(b *testing.B) {
		e := goke.CreateEntity(ecs)
		*goke.EnsureComponent[Position3](ecs, e, posDesc) = Position3{X: 1, Y: 2, Z: 3}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Uses the provided compInfo to jump straight to the correct column
			pos, err := goke.GetComponent[Position3](ecs, e, posDesc)
			if err == nil {
				pos.X += 1.0
			}
		}
	})
}

func BenchmarkEngine_Query(b *testing.B) {
	ecs := goke.New(
		goke.WithInitialEntityCap(100000),
		goke.WithDefaultArchetypeChunkSize(1024),
		goke.WithInitialArchetypeRegistryCap(128),
		goke.WithFreeIndicesCap(100000),
		goke.WithViewRegistryInitCap(64),
	)
	posType := goke.RegisterComponent[Position3](ecs)

	count := 100000
	for i := 0; i < count; i++ {
		e := goke.CreateEntity(ecs)
		*goke.EnsureComponent[Position3](ecs, e, posType) = Position3{X: float64(i)}
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
		goke.WithDefaultArchetypeChunkSize(4096),
		goke.WithFreeIndicesCap(count),
	)
	posDesc := goke.RegisterComponent[Position3](ecs)

	// --- SETUP PHASE ---
	// Pre-create entities outside the timed loop to isolate the cost of removal
	entities := make([]goke.Entity, count)
	for j := 0; j < count; j++ {
		entities[j] = goke.CreateEntity(ecs)
		goke.EnsureComponent[Position3](ecs, entities[j], posDesc)
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
				entities[j] = goke.CreateEntity(ecs)
				goke.EnsureComponent[Position3](ecs, entities[j], posDesc)
			}
			b.StartTimer()
		}

		goke.RemoveEntity(ecs, entities[idx])
	}
}

func BenchmarkEngine_AddRemove_Stability(b *testing.B) {
	ecs := goke.New(goke.WithInitialEntityCap(1024))
	posDesc := goke.RegisterComponent[Position3](ecs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := goke.CreateEntity(ecs)
		*goke.EnsureComponent[Position3](ecs, e, posDesc) = Position3{X: 1}

		// Usuwamy co drugą, żeby wymusić swapowanie w archetypie
		if i%2 == 0 {
			goke.RemoveEntity(ecs, e)
		}
	}
}
