package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

func BenchmarkEngine_Structural(b *testing.B) {
	ecs := goke.New()

	// --- 2. ARCHETYPE TRANSITION (Evolution) ---
	b.Run("Add 2nd Component", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]uid.UID64, b.N)

		posDesc := goke.RegCompType[Pos](ecs)
		blueprintA := goke.NewFactory1[Vel](ecs)

		index := 0
		for chunk := range blueprintA.Create(b.N) {
			for _, e := range chunk.Entity {
				entities[index] = e
				index++
			}
		}
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				*goke.UpsertComp[Pos](ecs, entities[i], posDesc) = Pos{X: 10, Y: 10}
			}
		})
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add Tag", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]uid.UID64, b.N)

		tagDesc := goke.RegCompType[Tag](ecs)
		blueprintA := goke.NewFactory1[Vel](ecs)

		offset := 0
		for chunk := range blueprintA.Create(b.N) {
			n := copy(entities[offset:], chunk.Entity)
			offset += n
		}
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				goke.UpsertComp[Tag](ecs, entities[i], tagDesc)
			}
		})
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove Component", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]uid.UID64, b.N)

		posDesc := goke.RegCompType[Pos](ecs)
		blueprintA := goke.NewFactory1[Vel](ecs)

		offset := 0
		for chunk := range blueprintA.Create(b.N) {
			n := copy(entities[offset:], chunk.Entity)
			offset += n
		}
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				goke.RemoveComp(ecs, entities[i], posDesc)
			}
		})
	})

	// --- 5. COMPONENT ACCESS (GET) ---

	// Benchmark for GetComp: Uses reflection to find Meta.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component", func(b *testing.B) {
		goke.Reset(ecs)

		velDesc := goke.RegCompType[Vel](ecs)
		blueprintA := goke.NewFactory1[Vel](ecs)

		var e uid.UID64
		for chunk := range blueprintA.Create(1) {
			e = chunk.Entity[0]
			chunk.Comp1[0] = Vel{X: 1, Y: 2}
		}

		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				pos := goke.GetComp[Vel](ecs, e, velDesc)
				pos.X += 1.0
			}
		})
	})

	// Benchmark for GetComp: Uses reflection to find Meta.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component (Safe)", func(b *testing.B) {
		goke.Reset(ecs)

		velDesc := goke.RegCompType[Vel](ecs)
		blueprintA := goke.NewFactory1[Vel](ecs)

		var e uid.UID64
		for chunk := range blueprintA.Create(1) {
			e = chunk.Entity[0]
			chunk.Comp1[0] = Vel{X: 1, Y: 2}
		}

		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				pos, err := goke.SafeGetComp[Vel](ecs, e, velDesc)
				if err == nil {
					pos.X += 1.0
				}
			}
		})
	})

	// Benchmark for GetComp: Uses reflection to find Meta.
	// This is the "Convenience Path" - slower but easier to use.
	b.Run("Get Component via View.Filter", func(b *testing.B) {
		goke.Reset(ecs)

		_ = goke.RegCompType[Vel](ecs)
		blueprintA := goke.NewFactory1[Vel](ecs)

		var e uid.UID64
		for chunk := range blueprintA.Create(1) {
			e = chunk.Entity[0]
			chunk.Comp1[0] = Vel{X: 1, Y: 2}
		}

		var vel goke.Col[Vel]
		query := goke.NewView(ecs, vel.Track())
		arr := []uid.UID64{e}

		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				fit := query.Filter(arr)
				for fit.Next() {
					vel.At(fit).X += 1.0
				}
			}
		})
	})
}

func BenchmarkEngine_RemoveEntity_Clean(b *testing.B) {
	count := 100000
	// Initialize the ecs with "Turbo" settings to pre-allocate memory buffers
	ecs := goke.New(
		goke.WithEntityCap(count),
		goke.WithEntityFreeCap(count),
	)
	_ = goke.RegCompType[Pos](ecs)

	// --- SETUP PHASE ---
	// Pre-create entities outside the timed loop to isolate the cost of removal
	blueprint := goke.NewFactory1[Pos](ecs)
	entities := make([]uid.UID64, count)
	offset := 0
	for chunk := range blueprint.Create(b.N) {
		n := copy(entities[offset:], chunk.Entity)
		offset += n
	}

	measurePerEntity(b, 1, func() {
		for i := 0; b.Loop(); i++ {
			idx := i % count

			// If b.N > count, we need to re-create the entity to ensure
			// we are benchmarking a real 'Remove' operation rather than
			// an early-exit check for a non-existent entity.
			if i >= count && i%count == 0 {
				b.StopTimer()
				offset := 0
				for chunk := range blueprint.Create(b.N) {
					n := copy(entities[offset:], chunk.Entity)
					offset += n
				}
				b.StartTimer()
			}

			goke.RemoveEntity(ecs, entities[idx])
		}
	})
}

func BenchmarkEngine_AddRemove_Stability(b *testing.B) {
	ecs := goke.New(goke.WithEntityCap(1024))
	_ = goke.RegCompType[Pos](ecs)
	blueprint := goke.NewFactory1[Pos](ecs)

	var e uid.UID64
	measurePerEntity(b, 1, func() {
		for i := 0; b.Loop(); i++ {
			for chunk := range blueprint.Create(1) {
				e = chunk.Entity[0]
				chunk.Comp1[0] = Pos{X: 1}
			}

			// Usuwamy co drugą, żeby wymusić swapowanie w archetypie
			if i%2 == 0 {
				goke.RemoveEntity(ecs, e)
			}
		}
	})
}
