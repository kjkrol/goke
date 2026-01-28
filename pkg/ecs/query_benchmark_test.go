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

func setupBenchmark(_ *testing.B, count int) (*Engine, []Entity) {
	eng := NewEngine()
	reg := eng.registry
	posTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Pos]())
	velTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Vel]())
	accTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Acc]())
	massTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Mass]())
	spinTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Spin]())
	charTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Char]())
	elecTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Elec]())
	magnTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Magn]())

	var entities []Entity
	for range count {
		e := reg.CreateEntity()
		reg.AssignByID(e, posTypeInfo, unsafe.Pointer(&Pos{1, 1}))
		reg.AssignByID(e, velTypeInfo, unsafe.Pointer(&Vel{1, 1}))
		reg.AssignByID(e, accTypeInfo, unsafe.Pointer(&Acc{1, 1}))
		reg.AssignByID(e, massTypeInfo, unsafe.Pointer(&Mass{}))
		reg.AssignByID(e, spinTypeInfo, unsafe.Pointer(&Spin{}))
		reg.AssignByID(e, charTypeInfo, unsafe.Pointer(&Char{}))
		reg.AssignByID(e, elecTypeInfo, unsafe.Pointer(&Elec{1}))
		reg.AssignByID(e, magnTypeInfo, unsafe.Pointer(&Magn{1}))
		entities = append(entities, e)
	}
	return eng, entities
}

// --- Benchmarki Standardowe (All) ---

func BenchmarkQuery0_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query := NewQuery0(eng)

	// The fn function is essential as it allows inlining logic and iteration, enabling faster reads using CPU L1/L2 Cache.
	fn := func() {
		for entity := range query.All() {
			entity.IsVirtual()
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkQuery1_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query1 := NewQuery1[Pos](eng)

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

func BenchmarkQuery2_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query2 := NewQuery2[Pos, Vel](eng)

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

func BenchmarkQuery3_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](eng)

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

func BenchmarkQuery3WithTag_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)

	query := NewQuery3[Pos, Vel, Acc](eng, WithTag[Mass]())

	fn := func() {
		for head := range query.All3() {
			p, v, a := head.V1, head.V2, head.V3
			p.X += v.X + a.X
		}
	}

	b.ResetTimer()
	for b.Loop() {
		fn()
	}
}

func BenchmarkQuery4_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query4 := NewQuery4[Pos, Vel, Acc, Mass](eng)

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

func BenchmarkQuery5_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query5 := NewQuery5[Pos, Vel, Acc, Mass, Spin](eng)

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

func BenchmarkQuery6_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query6 := NewQuery6[Pos, Vel, Acc, Mass, Spin, Char](eng)

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

func BenchmarkQuery7_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query7 := NewQuery7[Pos, Vel, Acc, Mass, Spin, Char, Elec](eng)

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

func BenchmarkQuery8_All(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query8 := NewQuery8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](eng)

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

func BenchmarkQuery3_Filtered100(b *testing.B) {
	eng, entities := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](eng)
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

func BenchmarkQuery1_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query1 := NewQuery1[Pos](eng)

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

func BenchmarkQuery2_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query2 := NewQuery2[Pos, Vel](eng)

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

func BenchmarkQuery3_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](eng)

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

func BenchmarkQuery4_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query4 := NewQuery4[Pos, Vel, Acc, Mass](eng)

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

func BenchmarkQuery5_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query5 := NewQuery5[Pos, Vel, Acc, Mass, Spin](eng)

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

func BenchmarkQuery6_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query6 := NewQuery6[Pos, Vel, Acc, Mass, Spin, Char](eng)

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

func BenchmarkQuery7_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query7 := NewQuery7[Pos, Vel, Acc, Mass, Spin, Char, Elec](eng)

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

func BenchmarkQuery8_PureAll(b *testing.B) {
	eng, _ := setupBenchmark(b, entitiesNumber)
	query8 := NewQuery8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](eng)

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

func BenchmarkQuery3_PureFiltered100(b *testing.B) {
	eng, entities := setupBenchmark(b, entitiesNumber)
	query3 := NewQuery3[Pos, Vel, Acc](eng)
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
