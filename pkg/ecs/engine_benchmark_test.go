package ecs_test

import (
	"testing"

	"github.com/kjkrol/goke/pkg/ecs"
)

type Position3 struct{ X, Y, Z float64 }
type Velocity3 struct{ X, Y, Z float64 }
type ProcessedTag struct{}

func BenchmarkEngine_Structural(b *testing.B) {
	engine := ecs.NewEngine(
		ecs.WithInitialEntityCap(100000),
		ecs.WithDefaultArchetypeChunkSize(4096),
		ecs.WithInitialArchetypeRegistryCap(128),
		ecs.WithFreeIndicesCap(100000),
		ecs.WithViewRegistryInitCap(64),
	)

	// Pre-registering for "Turbo" performance in benchmarks
	posInfo := ecs.RegisterComponent[Position3](engine)
	velInfo := ecs.RegisterComponent[Velocity3](engine)
	tagInfo := ecs.RegisterComponent[ProcessedTag](engine)

	// --- 1. ENTITY CREATION & INITIALIZATION ---
	// Tests CreateEntity
	b.Run("Create_Entity", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			e := engine.CreateEntity()
			_ = e
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Position} to {Position, Velocity}
	b.Run("Add_Component", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v, _ := ecs.AllocateComponentByInfo[Velocity3](engine, entities[i], velInfo)
			*v = Velocity3{X: 10, Y: 10}
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Position} to {Position, Velocity}
	b.Run("Add_2nd_Component", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
			p, _ := ecs.AllocateComponentByInfo[Position3](engine, entities[i], posInfo)
			*p = Position3{X: 1}
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v, _ := ecs.AllocateComponentByInfo[Velocity3](engine, entities[i], velInfo)
			*v = Velocity3{X: 10, Y: 10}
		}
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add_Tag", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ecs.AllocateComponentByInfo[ProcessedTag](engine, entities[i], tagInfo)
		}
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove_Component", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
			ecs.AllocateComponentByInfo[Position3](engine, entities[i], posInfo)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			engine.RemoveComponentByID(entities[i], posInfo)
		}
	})
}

func BenchmarkEngine_Query(b *testing.B) {
	engine := ecs.NewEngine(
		ecs.WithInitialEntityCap(100000),
		ecs.WithDefaultArchetypeChunkSize(1024),
		ecs.WithInitialArchetypeRegistryCap(128),
		ecs.WithFreeIndicesCap(100000),
		ecs.WithViewRegistryInitCap(64),
	)
	posInfo := ecs.RegisterComponent[Position3](engine)

	count := 100000
	for i := 0; i < count; i++ {
		e := engine.CreateEntity()
		pos, _ := ecs.AllocateComponentByInfo[Position3](engine, e, posInfo)
		*pos = Position3{X: float64(i)}
	}

	query := ecs.NewQuery1[Position3](engine)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for head := range query.All1() {
			head.V1.X += 0.1
		}
	}
}

func BenchmarkEngine_RemoveEntity_Clean(b *testing.B) {
	count := 100000
	// Initialize the engine with "Turbo" settings to pre-allocate memory buffers
	engine := ecs.NewEngine(
		ecs.WithInitialEntityCap(count),
		ecs.WithDefaultArchetypeChunkSize(4096),
		ecs.WithFreeIndicesCap(count),
	)
	posInfo := ecs.RegisterComponent[Position3](engine)

	// --- SETUP PHASE ---
	// Pre-create entities outside the timed loop to isolate the cost of removal
	entities := make([]ecs.Entity, count)
	for j := 0; j < count; j++ {
		entities[j] = engine.CreateEntity()
		ecs.AllocateComponentByInfo[Position3](engine, entities[j], posInfo)
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
				entities[j] = engine.CreateEntity()
				ecs.AllocateComponentByInfo[Position3](engine, entities[j], posInfo)
			}
			b.StartTimer()
		}

		engine.RemoveEntity(entities[idx])
	}
}

func BenchmarkEngine_AddRemove_Stability(b *testing.B) {
	engine := ecs.NewEngine(ecs.WithInitialEntityCap(1024))
	posInfo := ecs.RegisterComponent[Position3](engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := engine.CreateEntity()
		p, _ := ecs.AllocateComponentByInfo[Position3](engine, e, posInfo)
		*p = Position3{X: 1}

		// Usuwamy co drugą, żeby wymusić swapowanie w archetypie
		if i%2 == 0 {
			engine.RemoveEntity(e)
		}
	}
}
