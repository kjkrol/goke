package ecs

import (
	"reflect"
	"testing"
	"unsafe"
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

func setupBenchmark(_ *testing.B, count int) (*Registry, []Entity) {
	reg := NewRegistry()
	posTypeId := reg.RegisterComponentType(reflect.TypeFor[Pos]())
	velTypeId := reg.RegisterComponentType(reflect.TypeFor[Vel]())
	accTypeId := reg.RegisterComponentType(reflect.TypeFor[Acc]())
	massTypeId := reg.RegisterComponentType(reflect.TypeFor[Mass]())
	spinTypeId := reg.RegisterComponentType(reflect.TypeFor[Spin]())
	charTypeId := reg.RegisterComponentType(reflect.TypeFor[Char]())
	elecTypeId := reg.RegisterComponentType(reflect.TypeFor[Elec]())
	magnTypeId := reg.RegisterComponentType(reflect.TypeFor[Magn]())

	var entities []Entity
	for range count {
		e := reg.CreateEntity()
		reg.AssignByID(e, posTypeId, unsafe.Pointer(&Pos{1, 1}))
		reg.AssignByID(e, velTypeId, unsafe.Pointer(&Vel{1, 1}))
		reg.AssignByID(e, accTypeId, unsafe.Pointer(&Acc{1, 1}))
		reg.AssignByID(e, massTypeId, unsafe.Pointer(&Mass{}))
		reg.AssignByID(e, spinTypeId, unsafe.Pointer(&Spin{}))
		reg.AssignByID(e, charTypeId, unsafe.Pointer(&Char{}))
		reg.AssignByID(e, elecTypeId, unsafe.Pointer(&Elec{1}))
		reg.AssignByID(e, magnTypeId, unsafe.Pointer(&Magn{1}))
		entities = append(entities, e)
	}
	return reg, entities
}

// --- Benchmarki Standardowe (All) ---

func BenchmarkView1_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query1 := NewQuery1[Pos](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query1.All1() {
			pos := head.V1
			pos.X += pos.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView2_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query2 := NewQuery2[Pos, Vel](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query2.All2() {
			pos, vel := head.V1, head.V2
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView3_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query3.All3() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView4_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query4 := NewQuery4[Pos, Vel, Acc, Mass](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range query4.All4() {
			pos, vel, acc := head.V1, head.V2, head.V3
			_ = tail.V4 // Mass
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView5_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query5 := NewQuery5[Pos, Vel, Acc, Mass, Spin](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, tail := range query5.All5() {
			pos, vel, acc := head.V1, head.V2, head.V3
			_ = tail.V4
			_ = tail.V5
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView6_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query6 := NewQuery6[Pos, Vel, Acc, Mass, Spin, Char](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range query6.All6() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView7_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query7 := NewQuery7[Pos, Vel, Acc, Mass, Spin, Char, Elec](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range query7.All7() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView8_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query8 := NewQuery8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range query8.All8() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

// --- Benchmarki Filtrowane ---

func BenchmarkView3_Filtered100(b *testing.B) {
	reg, entities := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](reg)
	subset := entities[:100]

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query3.Filter3(subset) {
			pos, vel, acc := head.V1, head.V2, head.V3
			acc.X += vel.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

// --- Benchmarki Pure All ---

func BenchmarkView1_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query1 := NewQuery1[Pos](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query1.PureAll1() {
			pos := head.V1
			pos.X += pos.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView2_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query2 := NewQuery2[Pos, Vel](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query2.PureAll2() {
			pos, vel := head.V1, head.V2
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView3_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query3.PureAll3() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView4_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query4 := NewQuery4[Pos, Vel, Acc, Mass](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query4.PureAll4() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView5_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query5 := NewQuery5[Pos, Vel, Acc, Mass, Spin](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range query5.PureAll5() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView6_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query6 := NewQuery6[Pos, Vel, Acc, Mass, Spin, Char](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range query6.PureAll6() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView7_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query7 := NewQuery7[Pos, Vel, Acc, Mass, Spin, Char, Elec](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range query7.PureAll7() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkView8_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	query8 := NewQuery8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](reg)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head, _ := range query8.PureAll8() {
			pos, vel, acc := head.V1, head.V2, head.V3
			vel.X += acc.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

// --- Benchmarki Pure Filtered ---

func BenchmarkView3_PureFiltered100(b *testing.B) {
	reg, entities := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](reg)
	subset := entities[:100]

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for head := range query3.PureFilter3(subset) {
			pos, vel, acc := head.V1, head.V2, head.V3
			acc.X += vel.X
			pos.X += vel.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}
