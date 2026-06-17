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
		fcVelOpt := goke.Track(new(goke.Col[Vel]))
		blueprintA := goke.CreateEntFactory(ecs, fcVelOpt)

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
				*goke.UpsertComp[Pos](ecs, entities[i], posDesc) = Pos{X: 10, Y: 10}
			}
		})
	})

	// --- 3. TAGGING (Zero-size components) ---
	b.Run("Add Tag", func(b *testing.B) {
		goke.Reset(ecs)
		entities := make([]uid.UID64, b.N)

		tagDesc := goke.RegCompType[Tag](ecs)
		fcVelOpt := goke.Track(new(goke.Col[Vel]))
		blueprintA := goke.CreateEntFactory(ecs, fcVelOpt)

		offset := 0
		blueprintA.Create(b.N)
		for blueprintA.Next() {
			n := copy(entities[offset:], blueprintA.IDs)
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
		fcVelOpt := goke.Track(new(goke.Col[Vel]))
		blueprintA := goke.CreateEntFactory(ecs, fcVelOpt)

		offset := 0
		blueprintA.Create(b.N)
		for blueprintA.Next() {
			n := copy(entities[offset:], blueprintA.IDs)
			offset += n
		}
		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				goke.RemoveComp(ecs, entities[i], posDesc)
			}
		})
	})

	// --- 5. COMPONENT ACCESS (GET) ---

	b.Run("Get Component", func(b *testing.B) {
		goke.Reset(ecs)

		velDesc := goke.RegCompType[Vel](ecs)
		var vel goke.Col[Vel]
		blueprintA := goke.CreateEntFactory(ecs, goke.Track(&vel))

		blueprintA.Create(1)
		blueprintA.Next()
		e := blueprintA.IDs[0]
		vel.Slice(&blueprintA.Cursor)[0] = Vel{X: 1, Y: 2}

		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				pos := goke.GetComp[Vel](ecs, e, velDesc)
				pos.X += 1.0
			}
		})
	})

	b.Run("Get Component (Safe)", func(b *testing.B) {
		goke.Reset(ecs)

		velDesc := goke.RegCompType[Vel](ecs)
		var vel goke.Col[Vel]
		blueprintA := goke.CreateEntFactory(ecs, goke.Track(&vel))

		blueprintA.Create(1)
		blueprintA.Next()
		e := blueprintA.IDs[0]
		vel.Slice(&blueprintA.Cursor)[0] = Vel{X: 1, Y: 2}

		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				if pos := goke.GetComp[Vel](ecs, e, velDesc); pos != nil {
					pos.X += 1.0
				}
			}
		})
	})

	b.Run("Get Component via View.Filter", func(b *testing.B) {
		goke.Reset(ecs)

		_ = goke.RegCompType[Vel](ecs)
		var vel goke.Col[Vel]
		blueprintA := goke.CreateEntFactory(ecs, goke.Track(&vel))

		blueprintA.Create(1)
		blueprintA.Next()
		e := blueprintA.IDs[0]
		vel.Slice(&blueprintA.Cursor)[0] = Vel{X: 1, Y: 2}

		query := goke.CreateView(ecs, goke.Track(&vel))
		arr := []uid.UID64{e}

		measurePerEntity(b, 1, func() {
			for i := 0; i < b.N; i++ {
				fit := query.Filter(arr)
				for fit.Next() {
					vel.At(&fit.Cursor).X += 1.0
				}
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
	_ = goke.RegCompType[Pos](ecs)

	fcPosOpt := goke.Track(new(goke.Col[Pos]))
	blueprint := goke.CreateEntFactory(ecs, fcPosOpt)
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
	_ = goke.RegCompType[Pos](ecs)
	var pos goke.Col[Pos]
	blueprint := goke.CreateEntFactory(ecs, goke.Track(&pos))
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
