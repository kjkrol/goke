package bench

import (
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/internal/core"
)

const entitiesNumber = 1000

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }
type Mass struct{ M float32 }
type Spin struct{ S [32]float32 }
type Char struct{ V [32]float32 }
type Elec struct{ V float64 }
type Magn struct{ V float64 }
type T09 struct{ V float64 }
type T10 struct{ V float64 }

func setupBenchmark(count int) (*goke.ECS, []core.Entity) {
	ecs := goke.New()
	_ = goke.RegisterComponent[Pos](ecs)
	_ = goke.RegisterComponent[Vel](ecs)
	_ = goke.RegisterComponent[Acc](ecs)
	_ = goke.RegisterComponent[Mass](ecs)
	_ = goke.RegisterComponent[Spin](ecs)
	_ = goke.RegisterComponent[Char](ecs)
	_ = goke.RegisterComponent[Elec](ecs)
	_ = goke.RegisterComponent[Magn](ecs)
	_ = goke.RegisterComponent[T09](ecs)
	_ = goke.RegisterComponent[T10](ecs)

	var entities []core.Entity
	blueprint := goke.NewBlueprint10[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn, T09, T10](ecs)
	for range count {
		item := blueprint.Create()
		e, pos, vel, acc, mass, spin, char, elec, magn, v09, v10 := item.Entity, item.Comp1, item.Comp2, item.Comp3, item.Comp4, item.Comp5, item.Comp6, item.Comp7, item.Comp8, item.Comp9, item.Comp10

		*pos = Pos{rand.Float32() * 100, rand.Float32() * 100}
		*vel = Vel{rand.Float32() * 40, 1}
		*acc = Acc{rand.Float32(), 0.1}
		*mass = Mass{}
		*spin = Spin{}
		*char = Char{}
		*elec = Elec{1}
		*magn = Magn{1}
		*v09 = T09{}
		*v10 = T10{}

		entities = append(entities, e)
	}
	return ecs, entities
}

// --- Benchmark All ---

var GlobalCount int

func BenchmarkView0_All(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view := goke.NewView0(ecs)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		count := 0
		for entity := range view.All() {
			_ = entity
			count++
		}
		GlobalCount = count
	}

	if GlobalCount != entitiesNumber {
		b.Fatalf("View0 sanity check failed: expected 1000, got %d", GlobalCount)
	}
}

func BenchmarkView1_All(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view1 := goke.NewView1[Pos](ecs)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		count := 0
		for item := range view1.All() {
			pos := item.Comp1
			pos.X += pos.X
			count++
		}
		GlobalCount = count
	}

	if GlobalCount != entitiesNumber {
		b.Fatalf("View1 sanity check failed: expected 1000, got %d", GlobalCount)
	}
}

func BenchmarkView2_All(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view2 := goke.NewView2[Pos, Vel](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for item := range view2.All() {
			pos, vel := item.Comp1, item.Comp2
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
	ecs, _ := setupBenchmark(entitiesNumber)
	view3 := goke.NewView3[Pos, Vel, Acc](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for item := range view3.All() {
			pos, vel, acc := item.Comp1, item.Comp2, item.Comp3
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
	ecs, _ := setupBenchmark(entitiesNumber)
	view := goke.NewView3[Pos, Vel, Acc](ecs, goke.Include[Mass]())

	fn := func() {
		for item := range view.All() {
			p, v, a := item.Comp1, item.Comp2, item.Comp3
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
	ecs, _ := setupBenchmark(entitiesNumber)
	view4 := goke.NewView4[Pos, Vel, Acc, Mass](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range view4.All() {
			pos, vel, acc, _ := head.Comp1, head.Comp2, head.Comp3, tail.Comp4
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
	ecs, _ := setupBenchmark(entitiesNumber)
	view5 := goke.NewView5[Pos, Vel, Acc, Mass, Spin](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range view5.All() {
			pos, vel, acc, _, _ := head.Comp1, head.Comp2, head.Comp3, tail.Comp4, tail.Comp5
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
	ecs, _ := setupBenchmark(entitiesNumber)
	view6 := goke.NewView6[Pos, Vel, Acc, Mass, Spin, Char](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for item := range view6.All() {
			pos, vel, acc := item.Comp1, item.Comp2, item.Comp3
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
	ecs, _ := setupBenchmark(entitiesNumber)
	view7 := goke.NewView7[Pos, Vel, Acc, Mass, Spin, Char, Elec](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range view7.All() {
			pos, vel, acc, char := head.Comp1, head.Comp2, head.Comp3, tail.Comp6
			vel.X += acc.X
			pos.X += vel.X
			_ = char
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView8_All(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view8 := goke.NewView8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for item := range view8.All() {
			pos, vel, acc := item.Comp1, item.Comp2, item.Comp3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView9_All(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view10 := goke.NewView9[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn, T09](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for item := range view10.All() {
			pos, vel, acc := item.Comp1, item.Comp2, item.Comp3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView10_All(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view10 := goke.NewView10[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn, T09, T10](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range view10.All() {
			pos, vel, acc := head.Comp1, head.Comp2, head.Comp3
			m, s, c, e, mag, v09, v10 := tail.Comp4, tail.Comp5, tail.Comp6, tail.Comp7, tail.Comp8, tail.Comp9, tail.Comp10
			vel.X += acc.X
			pos.X += vel.X
			m.M = 1
			s.S[0] = 1
			c.V[0] = 1
			e.V = 1
			mag.V = 1
			v09.V = 1
			v10.V = 1
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// --- Benchmark Filter ---

func BenchmarkView0_Filter100(b *testing.B) {
	b.StopTimer()
	ecs, entities := setupBenchmark(entitiesNumber)
	view3 := goke.NewView0(ecs)
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

func BenchmarkView2_Filter100(b *testing.B) {
	b.StopTimer()
	ecs, entities := setupBenchmark(entitiesNumber)
	view3 := goke.NewView2[Pos, Vel](ecs)
	subset := entities[:100]

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for item := range view3.Filter(subset) {
			pos, vel := item.Comp1, item.Comp2
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView3_Filter100(b *testing.B) {
	b.StopTimer()
	ecs, entities := setupBenchmark(entitiesNumber)
	view3 := goke.NewView3[Pos, Vel, Acc](ecs)
	subset := entities[:100]

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for item := range view3.Filter(subset) {
			pos, vel, acc := item.Comp1, item.Comp2, item.Comp3
			acc.X += vel.X
			pos.X += vel.X
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// --- Benchmark Each ---

func BenchmarkView1_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view1 := goke.NewView1[Pos](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view1.Each(func(entities []core.Entity, c1 []Pos) {
			for i := range len(entities) {
				pos := c1[i]
				pos.X += 1
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView2_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view2 := goke.NewView2[Pos, Vel](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view2.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel) {
			for i := range len(entities) {
				pos, vel := c1[i], c2[i]
				vel.X += 1
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView3_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view3 := goke.NewView3[Pos, Vel, Acc](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view3.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc) {
			for i := range len(entities) {
				pos, vel, acc := c1[i], c2[i], c3[i]
				vel.X += acc.X
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView4_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view4 := goke.NewView4[Pos, Vel, Acc, Mass](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view4.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc, c4 []Mass) {
			for i := range len(entities) {
				pos, vel, acc, _ := c1[i], c2[i], c3[i], c4[i]
				vel.X += acc.X
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView5_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view5 := goke.NewView5[Pos, Vel, Acc, Mass, Spin](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view5.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc, c4 []Mass, c5 []Spin) {
			for i := range len(entities) {
				pos, vel, acc, _, _ := c1[i], c2[i], c3[i], c4[i], c5[i]
				vel.X += acc.X
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView6_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view6 := goke.NewView6[Pos, Vel, Acc, Mass, Spin, Char](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view6.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc, c4 []Mass, c5 []Spin, c6 []Char) {
			for i := range len(entities) {
				pos, vel, acc, _, _, _ := c1[i], c2[i], c3[i], c4[i], c5[i], c6[i]
				vel.X += acc.X
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView7_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view7 := goke.NewView7[Pos, Vel, Acc, Mass, Spin, Char, Elec](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view7.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc, c4 []Mass, c5 []Spin, c6 []Char, c7 []Elec) {
			for i := range len(entities) {
				pos, vel, acc, _, _, _, _ := c1[i], c2[i], c3[i], c4[i], c5[i], c6[i], c7[i]
				vel.X += acc.X
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView8_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view8 := goke.NewView8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view8.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc, c4 []Mass, c5 []Spin, c6 []Char, c7 []Elec, c8 []Magn) {
			for i := range len(entities) {
				pos, vel, acc, _, _, _, _, _ := c1[i], c2[i], c3[i], c4[i], c5[i], c6[i], c7[i], c8[i]
				vel.X += acc.X
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView9_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view9 := goke.NewView9[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn, T09](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view9.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc, c4 []Mass, c5 []Spin, c6 []Char, c7 []Elec, c8 []Magn, c9 []T09) {
			for i := range len(entities) {
				pos, vel, acc, _, _, _, _, _, _ := c1[i], c2[i], c3[i], c4[i], c5[i], c6[i], c7[i], c8[i], c9[i]
				vel.X += acc.X
				pos.X += vel.X
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkView10_Each(b *testing.B) {
	b.StopTimer()
	ecs, _ := setupBenchmark(entitiesNumber)
	view10 := goke.NewView10[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn, T09, T10](ecs)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		view10.Each(func(entities []core.Entity, c1 []Pos, c2 []Vel, c3 []Acc, c4 []Mass, c5 []Spin, c6 []Char, c7 []Elec, c8 []Magn, c9 []T09, c10 []T10) {
			for i := range len(entities) {
				pos, vel, acc, m, s, c, e, mag, v09, v10 := c1[i], c2[i], c3[i], c4[i], c5[i], c6[i], c7[i], c8[i], c9[i], c10[i]
				vel.X += acc.X
				pos.X += vel.X
				m.M = 1
				s.S[0] = 1
				c.V[0] = 1
				e.V = 1
				mag.V = 1
				v09.V = 1
				v10.V = 1
			}
		})
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}
