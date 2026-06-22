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

		posID := goke.RegComp[Pos](ecs)
		fcVelOpt := goke.Track(new(goke.Col[Vel]))
		blueprintA := goke.CreateFactory(ecs, fcVelOpt)

		index := 0
		blueprintA.Create(b.N)
		for blueprintA.Next() {
			for _, e := range blueprintA.IDs {
				entities[index] = e
				index++
			}
		}
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				*goke.UpsertComp[Pos](ecs, entities[i], posID) = Pos{X: 10, Y: 10}
			}
		})
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add Tag", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]uid.UID64, b.N)

		tagID := goke.RegComp[Tag](ecs)
		fcVelOpt := goke.Track(new(goke.Col[Vel]))
		blueprintA := goke.CreateFactory(ecs, fcVelOpt)

		offset := 0
		blueprintA.Create(b.N)
		for blueprintA.Next() {
			n := copy(entities[offset:], blueprintA.IDs)
			offset += n
		}
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				goke.UpsertComp[Tag](ecs, entities[i], tagID)
			}
		})
	})

	// --- 4. COMPONENT REMOVAL ---
	b.Run("Remove Component", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]uid.UID64, b.N)

		posID := goke.RegComp[Pos](ecs)
		fcVelOpt := goke.Track(new(goke.Col[Vel]))
		blueprintA := goke.CreateFactory(ecs, fcVelOpt)

		offset := 0
		blueprintA.Create(b.N)
		for blueprintA.Next() {
			n := copy(entities[offset:], blueprintA.IDs)
			offset += n
		}
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				goke.RemoveComp(ecs, entities[i], posID)
			}
		})
	})
}

func BenchmarkEngine_RemoveEntity_Clean(b *testing.B) {
	count := 100000
	ecs := goke.New(
		goke.WithEntityCap(count),
		goke.WithEntityFreeCap(count),
	)
	_ = goke.RegComp[Pos](ecs)

	fcPosOpt := goke.Track(new(goke.Col[Pos]))
	blueprint := goke.CreateFactory(ecs, fcPosOpt)
	entities := make([]uid.UID64, count)

	refill := func() {
		offset := 0
		blueprint.Create(b.N)
		for blueprint.Next() {
			n := copy(entities[offset:], blueprint.IDs)
			offset += n
		}
	}
	refill()

	measurePerEntity(b, 1, func() {
		for i := 0; b.Loop(); i++ {
			idx := i % count

			if i >= count && i%count == 0 {
				b.StopTimer()
				refill()
				b.StartTimer()
			}

			goke.RemoveEnt(ecs, entities[idx])
		}
	})
}

func BenchmarkEngine_AddRemove_Stability(b *testing.B) {
	ecs := goke.New(goke.WithEntityCap(1024))
	_ = goke.RegComp[Pos](ecs)
	var pos goke.Col[Pos]
	blueprint := goke.CreateFactory(ecs, goke.Track(&pos))
	fc := &blueprint.Cursor

	var e uid.UID64
	measurePerEntity(b, 1, func() {
		for i := 0; b.Loop(); i++ {
			blueprint.Create(1)
			blueprint.Next()
			e = blueprint.IDs[0]
			pos.Slice(fc)[0] = Pos{X: 1}

			if i%2 == 0 {
				goke.RemoveEnt(ecs, e)
			}
		}
	})
}
