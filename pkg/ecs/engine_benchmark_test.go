package ecs_test

import (
	"testing"

	"github.com/kjkrol/goke/pkg/ecs"
)

type Position3 struct{ X, Y, Z float64 }
type Velocity3 struct{ X, Y, Z float64 }
type ProcessedTag struct{}

func BenchmarkEngine_Structural(b *testing.B) {
	engine := ecs.NewEngine()

	// Pre-registering for "Turbo" performance in benchmarks
	posInfo := ecs.RegisterComponent[Position3](engine)
	velInfo := ecs.RegisterComponent[Velocity3](engine)
	tagInfo := ecs.RegisterComponent[ProcessedTag](engine)

	// --- 1. ENTITY CREATION & INITIALIZATION ---
	// Tests AddComponentByInfo + In-place dereference (*ptr = value)
	b.Run("Create_And_Init_Entity", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			e := engine.CreateEntity()
			pos, _ := ecs.AddComponentByInfo[Position3](engine, e, posInfo)
			*pos = Position3{X: 1, Y: 2, Z: 3}
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Position} to {Position, Velocity}
	b.Run("Add_Component_Transition", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
			p, _ := ecs.AddComponentByInfo[Position3](engine, entities[i], posInfo)
			*p = Position3{X: 1}
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v, _ := ecs.AddComponentByInfo[Velocity3](engine, entities[i], velInfo)
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
			ecs.AddComponentByInfo[ProcessedTag](engine, entities[i], tagInfo)
		}
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove_Component", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
			ecs.AddComponentByInfo[Position3](engine, entities[i], posInfo)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			engine.RemoveComponentByID(entities[i], posInfo)
		}
	})
}

func BenchmarkEngine_Query(b *testing.B) {
	engine := ecs.NewEngine()
	posInfo := ecs.RegisterComponent[Position3](engine)

	count := 100000
	for i := 0; i < count; i++ {
		e := engine.CreateEntity()
		pos, _ := ecs.AddComponentByInfo[Position3](engine, e, posInfo)
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

func BenchmarkEngine_MassOperations(b *testing.B) {
	// SCENARIO: Remove every second entity to stress Swap & Pop
	b.Run("Fragmentation_Stress", func(b *testing.B) {
		count := 100000
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			engine := ecs.NewEngine()
			posInfo := ecs.RegisterComponent[Position3](engine)
			entities := make([]ecs.Entity, count)
			for j := 0; j < count; j++ {
				entities[j] = engine.CreateEntity()
				ecs.AddComponentByInfo[Position3](engine, entities[j], posInfo)
			}
			b.StartTimer()

			for j := 0; j < count; j += 2 {
				engine.RemoveEntity(entities[j])
			}
		}
	})
}
