package bench

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/internal/core"
)

const entitiesNumber = 1000

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }
type Mass struct{} // TAG
type Spin struct{ S [32]float32 }
type Char struct{ V [32]float32 }
type Elec struct{ V float64 }
type Magn struct{ V float64 }

func setupBenchmark(_ *testing.B, count int) (*goke.Engine, []core.Entity) {
	eng := goke.NewEngine()
	posType := goke.ComponentRegister[Pos](eng)
	velType := goke.ComponentRegister[Vel](eng)
	accType := goke.ComponentRegister[Acc](eng)
	massType := goke.ComponentRegister[Mass](eng)
	spinType := goke.ComponentRegister[Spin](eng)
	charType := goke.ComponentRegister[Char](eng)
	elecType := goke.ComponentRegister[Elec](eng)
	magnType := goke.ComponentRegister[Magn](eng)

	var entities []core.Entity
	for range count {
		e := goke.EntityCreate(eng)

		if pos, err := goke.EntityEnsureComponent[Pos](eng, e, posType); err == nil {
			*(*Pos)(pos) = Pos{1, 1}
		}
		if vel, err := goke.EntityEnsureComponent[Vel](eng, e, velType); err == nil {
			*(*Vel)(vel) = Vel{1, 1}
		}
		if acc, err := goke.EntityEnsureComponent[Acc](eng, e, accType); err == nil {
			*(*Acc)(acc) = Acc{1, 1}
		}
		if mass, err := goke.EntityEnsureComponent[Mass](eng, e, massType); err == nil {
			*(*Mass)(mass) = Mass{}
		}
		if spin, err := goke.EntityEnsureComponent[Spin](eng, e, spinType); err == nil {
			*(*Spin)(spin) = Spin{}
		}
		if char, err := goke.EntityEnsureComponent[Char](eng, e, charType); err == nil {
			*(*Char)(char) = Char{}
		}
		if elec, err := goke.EntityEnsureComponent[Elec](eng, e, elecType); err == nil {
			*(*Elec)(elec) = Elec{1}
		}
		if magn, err := goke.EntityEnsureComponent[Magn](eng, e, magnType); err == nil {
			*(*Magn)(magn) = Magn{1}
		}
		entities = append(entities, e)
	}
	return eng, entities
}

// --- Benchmarki Standardowe (All) ---

func BenchmarkView0_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view := goke.NewView0(eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for entity := range view.All() {
			entity.IsVirtual()
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView1_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view1 := goke.NewView1[Pos](eng)

	fn := func() {
		for head := range view1.All() {
			pos := head.V1
			pos.X += pos.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView2_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view2 := goke.NewView2[Pos, Vel](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view2.All() {
			pos, vel := head.V1, head.V2
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView3_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view3 := goke.NewView3[Pos, Vel, Acc](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view3.All() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView3WithTag_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view := goke.NewView3[Pos, Vel, Acc](eng, core.WithTag[Mass]())

	fn := func() {
		for head := range view.All() {
			p, v, a := head.V1, head.V2, head.V3
			p.X += v.X + a.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView4_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view4 := goke.NewView4[Pos, Vel, Acc, Mass](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range view4.All() {
			pos, vel, acc := head.V1, head.V2, head.V3
			_ = tail.V4 // Mass
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView5_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view5 := goke.NewView5[Pos, Vel, Acc, Mass, Spin](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range view5.All() {
			pos, vel, acc := head.V1, head.V2, head.V3
			_ = tail.V4
			_ = tail.V5
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView6_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view6 := goke.NewView6[Pos, Vel, Acc, Mass, Spin, Char](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range view6.All() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView7_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view7 := goke.NewView7[Pos, Vel, Acc, Mass, Spin, Char, Elec](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range view7.All() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView8_All(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view8 := goke.NewView8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range view8.All() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// --- Benchmarki Filtrowane ---

func BenchmarkView0_Filter100(b *testing.B) {
	b.StopTimer()
	eng, entities := setupBenchmark(b, entitiesNumber)
	view3 := goke.NewView0(eng)
	subset := entities[:100]

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for entity := range view3.Filter(subset) {
			_ = entity
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView3_Filter100(b *testing.B) {
	b.StopTimer()
	eng, entities := setupBenchmark(b, entitiesNumber)
	view3 := goke.NewView3[Pos, Vel, Acc](eng)
	subset := entities[:100]

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view3.Filter(subset) {
			pos, vel, acc := head.V1, head.V2, head.V3
			acc.X += vel.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// --- Benchmarki Pure All ---

func BenchmarkView1_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view1 := goke.NewView1[Pos](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view1.Values() {
			pos := head.V1
			pos.X += pos.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView2_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view2 := goke.NewView2[Pos, Vel](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view2.Values() {
			pos, vel := head.V1, head.V2
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView3_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view3 := goke.NewView3[Pos, Vel, Acc](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view3.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView4_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view4 := goke.NewView4[Pos, Vel, Acc, Mass](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view4.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView5_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view5 := goke.NewView5[Pos, Vel, Acc, Mass, Spin](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range view5.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView6_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view6 := goke.NewView6[Pos, Vel, Acc, Mass, Spin, Char](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range view6.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView7_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view7 := goke.NewView7[Pos, Vel, Acc, Mass, Spin, Char, Elec](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range view7.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView8_Values(b *testing.B) {
	b.StopTimer()
	eng, _ := setupBenchmark(b, entitiesNumber)
	view8 := goke.NewView8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range view8.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// --- Benchmarki FilterValues ---

func BenchmarkView3_FilterValues100(b *testing.B) {
	b.StopTimer()
	eng, entities := setupBenchmark(b, entitiesNumber)
	view3 := goke.NewView3[Pos, Vel, Acc](eng)
	subset := entities[:100]

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range view3.FilterValues(subset) {
			pos, vel, acc := head.V1, head.V2, head.V3
			acc.X += vel.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}
