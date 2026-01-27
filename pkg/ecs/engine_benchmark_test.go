package ecs_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/pkg/ecs"
)

type Position3 struct{ X, Y, Z float64 }
type Velocity3 struct{ X, Y, Z float64 }
type ProcessedTag struct{}

func BenchmarkEngine_Structural_RealData(b *testing.B) {
	engine := ecs.NewEngine()

	posInfo := engine.RegisterComponentType(reflect.TypeFor[Position3]())
	velInfo := engine.RegisterComponentType(reflect.TypeFor[Velocity3]())
	tagInfo := engine.RegisterComponentType(reflect.TypeFor[ProcessedTag]())

	// --- WARMUP ---
	// Create and remove many entities to force slices and maps to grow to
	// a stable size before we start measuring.
	warmupCount := 100000
	tempEntities := make([]ecs.Entity, warmupCount)
	for i := 0; i < warmupCount; i++ {
		tempEntities[i] = engine.CreateEntity()
		engine.AllocateComponentMemoryByID(tempEntities[i], posInfo)
	}
	for i := 0; i < warmupCount; i++ {
		engine.RemoveEntity(tempEntities[i])
	}
	// Now internal buffers are large enough.

	// --- 1. DATA COMPONENT ASSIGNMENT (Standard - memmove) ---
	b.Run("Assign_Data_Component", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := Position3{X: float64(i), Y: float64(i), Z: float64(i)}
			engine.AssignByID(entities[i], posInfo, unsafe.Pointer(&p))
		}
	})

	// --- 2. DATA COMPONENT ALLOCATE (In-Place - No Escape) ---
	// This should result in 0 B/op because we don't pass a pointer to a local variable.
	b.Run("Allocate_Data_Component", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Get memory directly from the archetype
			ptr, _ := engine.AllocateComponentMemoryByID(entities[i], posInfo)
			p := (*Position3)(ptr)
			// Write directly to the destination
			p.X = float64(i)
			p.Y = float64(i)
			p.Z = float64(i)
		}
	})

	// --- 3. ARCHETYPE TRANSITION (Standard) ---
	b.Run("Assign_Transition_Complex", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
			p := Position3{X: float64(i)}
			engine.AssignByID(entities[i], posInfo, unsafe.Pointer(&p))
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v := Velocity3{X: float64(i), Y: 1, Z: 1}
			engine.AssignByID(entities[i], velInfo, unsafe.Pointer(&v))
		}
	})

	// --- 4. ARCHETYPE TRANSITION (In-Place) ---
	b.Run("Allocate_Transition_Complex", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
			// Using existing Assign for setup (or you could use Allocate here too)
			p := Position3{X: float64(i)}
			engine.AssignByID(entities[i], posInfo, unsafe.Pointer(&p))
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Structural change + get pointer to the new memory slot
			ptr, _ := engine.AllocateComponentMemoryByID(entities[i], velInfo)
			v := (*Velocity3)(ptr)
			v.X = float64(i)
			v.Y = 1
			v.Z = 1
		}
	})

	// --- 5. TAG ASSIGNMENT ---
	b.Run("Assign_Tag", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			engine.AssignByID(entities[i], tagInfo, nil)
		}
	})

	// --- 6. COMPONENT REMOVAL (Unassign) ---
	b.Run("Unassign_Component", func(b *testing.B) {
		entities := make([]ecs.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = engine.CreateEntity()
			p := Position3{X: float64(i)}
			engine.AssignByID(entities[i], posInfo, unsafe.Pointer(&p))
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			engine.UnassignByID(entities[i], posInfo)
		}
	})
}

func BenchmarkEngine_Query_RealData(b *testing.B) {
	engine := ecs.NewEngine()
	posInfo := engine.RegisterComponentType(reflect.TypeFor[Position3]())

	// Przygotujmy 100,000 encji
	count := 100000
	for i := 0; i < count; i++ {
		e := engine.CreateEntity()
		ptr, _ := engine.AllocateComponentMemoryByID(e, posInfo)
		p := (*Position3)(ptr)
		p.X, p.Y, p.Z = float64(i), 1.0, 2.0
	}

	// Twoje Query1
	query := ecs.NewQuery1[Position3](engine.Registry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Symulujemy system aktualizacji pozycji
		for head := range query.All1() {
			head.V1.X += 0.1 // Bezpośrednia modyfikacja przez wskaźnik
		}
	}
}

func BenchmarkEngine_Fragmentation_And_MassDelete(b *testing.B) {
	engine := ecs.NewEngine()
	posInfo := engine.RegisterComponentType(reflect.TypeFor[Position3]())

	// 1. Initial stable state setup (100k entities)
	count := 100000
	entities := make([]ecs.Entity, count)
	for i := 0; i < count; i++ {
		entities[i] = engine.CreateEntity()
		ptr, _ := engine.AllocateComponentMemoryByID(entities[i], posInfo)
		p := (*Position3)(ptr)
		p.X = float64(i)
	}

	// --- SCENARIO A: Delete every second entity (fragmentation test) ---
	// This forces continuous Swap & Pop operations within archetypes.
	b.Run("Delete_Every_Second_Entity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Fresh engine for each iteration to ensure clean state
			engine = ecs.NewEngine()
			posInfo = engine.RegisterComponentType(reflect.TypeFor[Position3]())
			for j := 0; j < count; j++ {
				entities[j] = engine.CreateEntity()
				engine.AllocateComponentMemoryByID(entities[j], posInfo)
			}
			b.StartTimer()

			// Remove every second entity to stress the Swap & Pop mechanism
			for j := 0; j < count; j += 2 {
				engine.RemoveEntity(entities[j])
			}
		}
	})

	// --- SCENARIO B: Query after mass deletion ---
	// Checks if Query performance remains optimal (compact memory) after 50% removal.
	b.Run("Query_After_Fragmentation", func(b *testing.B) {
		// Setup: Remove 50% of entities once before benchmarking
		for j := 0; j < count; j += 2 {
			engine.RemoveEntity(entities[j])
		}

		query := ecs.NewQuery1[Position3](engine.Registry)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			counter := 0
			for head := range query.All1() {
				head.V1.X += 1.0
				counter++
			}
			// Verify that the iterator only visits the remaining 50k entities
			if i == 0 && counter != count/2 {
				b.Errorf("Expected %d entities, found %d", count/2, counter)
			}
		}
	})
}
