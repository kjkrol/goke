package bench

import (
	"testing"

	"github.com/kjkrol/goke"
)

type Position3 struct{ X, Y, Z float64 }
type Velocity3 struct{ X, Y, Z float64 }
type ProcessedTag struct{}

func BenchmarkEngine_Structural(b *testing.B) {
	engine := goke.NewEngine(
		goke.WithInitialEntityCap(100000),
		goke.WithDefaultArchetypeChunkSize(4096),
		goke.WithInitialArchetypeRegistryCap(128),
		goke.WithFreeIndicesCap(100000),
		goke.WithViewRegistryInitCap(64),
	)

	// Pre-registering for "Turbo" performance in bench
	posType := goke.ComponentRegister[Position3](engine)
	velType := goke.ComponentRegister[Velocity3](engine)
	tagType := goke.ComponentRegister[ProcessedTag](engine)

	// initialize archetypes
	eStart1 := goke.EntityCreate(engine)
	goke.EntityEnsureComponent[Velocity3](engine, eStart1, velType)

	eStart2 := goke.EntityCreate(engine)
	goke.EntityEnsureComponent[Position3](engine, eStart2, posType)
	goke.EntityEnsureComponent[Velocity3](engine, eStart2, velType)

	eStart3 := goke.EntityCreate(engine)
	goke.EntityEnsureComponent[ProcessedTag](engine, eStart3, tagType)

	// --- 1. ENTITY CREATION & INITIALIZATION ---
	// Tests EntityCreate
	b.Run("Create_Entity", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			e := goke.EntityCreate(engine)
			_ = e
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Position} to {Position, Velocity}
	b.Run("Add_Component", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.EntityCreate(engine)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v, _ := goke.EntityEnsureComponent[Velocity3](engine, entities[i], velType)
			*v = Velocity3{X: 10, Y: 10}
		}
	})

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	// Tests moving an entity from {Position} to {Position, Velocity}
	b.Run("Add_2nd_Component", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.EntityCreate(engine)
			p, _ := goke.EntityEnsureComponent[Position3](engine, entities[i], posType)
			*p = Position3{X: 1}
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v, _ := goke.EntityEnsureComponent[Velocity3](engine, entities[i], velType)
			*v = Velocity3{X: 10, Y: 10}
		}
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add_Tag", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.EntityCreate(engine)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			goke.EntityEnsureComponent[ProcessedTag](engine, entities[i], tagType)
		}
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove_Component", func(b *testing.B) {
		entities := make([]goke.Entity, b.N)
		for i := 0; i < b.N; i++ {
			entities[i] = goke.EntityCreate(engine)
			goke.EntityEnsureComponent[Position3](engine, entities[i], posType)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			goke.EntityRemoveComponent(engine, entities[i], posType)
		}
	})

	// --- 5. COMPONENT ACCESS (GET) ---

	// Benchmark for ComponentGet: Uses reflection to find ComponentInfo.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get_Component_Reflect", func(b *testing.B) {
		e := goke.EntityCreate(engine)
		p, _ := goke.EntityEnsureComponent[Position3](engine, e, posType)
		*p = Position3{X: 1, Y: 2, Z: 3}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Internally calls reflect.TypeFor[T]() and registry lookups
			pos, err := goke.EntityGetComponent[Position3](engine, e, posType)
			if err == nil {
				pos.X += 1.0
			}
		}
	})

	// Benchmark for GetComponentByType: Uses pre-fetched ComponentInfo.
	// This is the "Fast Path" - bypasses reflection and registry maps.
	b.Run("Get_Component_Direct", func(b *testing.B) {
		e := goke.EntityCreate(engine)
		p, _ := goke.EntityEnsureComponent[Position3](engine, e, posType)
		*p = Position3{X: 1, Y: 2, Z: 3}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Uses the provided compInfo to jump straight to the correct column
			pos, err := goke.EntityGetComponent[Position3](engine, e, posType)
			if err == nil {
				pos.X += 1.0
			}
		}
	})
}

func BenchmarkEngine_Query(b *testing.B) {
	engine := goke.NewEngine(
		goke.WithInitialEntityCap(100000),
		goke.WithDefaultArchetypeChunkSize(1024),
		goke.WithInitialArchetypeRegistryCap(128),
		goke.WithFreeIndicesCap(100000),
		goke.WithViewRegistryInitCap(64),
	)
	posInfo := goke.ComponentRegister[Position3](engine)

	count := 100000
	for i := 0; i < count; i++ {
		e := goke.EntityCreate(engine)
		pos, _ := goke.EntityEnsureComponent[Position3](engine, e, posInfo)
		*pos = Position3{X: float64(i)}
	}

	view := goke.NewView1[Position3](engine)

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
	// Initialize the engine with "Turbo" settings to pre-allocate memory buffers
	engine := goke.NewEngine(
		goke.WithInitialEntityCap(count),
		goke.WithDefaultArchetypeChunkSize(4096),
		goke.WithFreeIndicesCap(count),
	)
	posInfo := goke.ComponentRegister[Position3](engine)

	// --- SETUP PHASE ---
	// Pre-create entities outside the timed loop to isolate the cost of removal
	entities := make([]goke.Entity, count)
	for j := 0; j < count; j++ {
		entities[j] = goke.EntityCreate(engine)
		goke.EntityEnsureComponent[Position3](engine, entities[j], posInfo)
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
				entities[j] = goke.EntityCreate(engine)
				goke.EntityEnsureComponent[Position3](engine, entities[j], posInfo)
			}
			b.StartTimer()
		}

		goke.EntityRemove(engine, entities[idx])
	}
}

func BenchmarkEngine_AddRemove_Stability(b *testing.B) {
	engine := goke.NewEngine(goke.WithInitialEntityCap(1024))
	posInfo := goke.ComponentRegister[Position3](engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := goke.EntityCreate(engine)
		p, _ := goke.EntityEnsureComponent[Position3](engine, e, posInfo)
		*p = Position3{X: 1}

		// Usuwamy co drugą, żeby wymusić swapowanie w archetypie
		if i%2 == 0 {
			goke.EntityRemove(engine, e)
		}
	}
}
