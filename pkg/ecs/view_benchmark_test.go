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

func view1All(v *View) {
	for head := range All1[Pos](v) {
		pos := head.V1
		pos.X += pos.X
	}
}

func view2All(v *View) {
	for head := range All2[Pos, Vel](v) {
		pos, vel := head.V1, head.V2
		pos.X += vel.X
	}
}

func view3All(v *View) {
	for head := range All3[Pos, Vel, Acc](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view4All(v *View) {
	for head := range All4[Pos, Vel, Acc, Mass](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		// acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view5All(v *View) {
	for head := range All5[Pos, Vel, Acc, Mass, Spin](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view6All(v *View) {
	for head := range All6[Pos, Vel, Acc, Mass, Spin, Char](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view7All(v *View) {
	for head := range All7[Pos, Vel, Acc, Mass, Spin, Char, Elec](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view8All(v *View) {
	for head := range All8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view3Filter(v *View, entities []Entity) {
	for head := range Filter3[Pos, Vel, Acc](v, entities) {
		pos, vel, acc := head.V1, head.V2, head.V3
		acc.X += vel.X
		pos.X += vel.X
	}
}

func BenchmarkView1_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view1 := NewView1[Pos](reg)

	b.ResetTimer()
	for b.Loop() {
		view1All(view1)
	}
}

func BenchmarkView2_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view2 := NewView2[Pos, Vel](reg)

	b.ResetTimer()
	for b.Loop() {
		view2All(view2)
	}
}

func BenchmarkView3_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view3 := NewView3[Pos, Vel, Acc](reg)

	b.ResetTimer()
	for b.Loop() {
		view3All(view3)
	}
}

func BenchmarkView4_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view4 := NewView4[Pos, Vel, Acc, Mass](reg)

	b.ResetTimer()
	for b.Loop() {
		view4All(view4)
	}
}

func BenchmarkView5_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view5 := NewView5[Pos, Vel, Acc, Mass, Spin](reg)

	b.ResetTimer()
	for b.Loop() {
		view5All(view5)
	}
}

func BenchmarkView6_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view6 := NewView6[Pos, Vel, Acc, Mass, Spin, Char](reg)

	b.ResetTimer()
	for b.Loop() {
		view6All(view6)
	}
}

func BenchmarkView7_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view7 := NewView7[Pos, Vel, Acc, Mass, Spin, Char, Elec](reg)

	b.ResetTimer()
	for b.Loop() {
		view7All(view7)
	}
}

func BenchmarkView8_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view8 := NewView8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](reg)

	b.ResetTimer()
	for b.Loop() {
		view8All(view8)
	}
}

func BenchmarkView3_Filtered100(b *testing.B) {
	reg, entities := setupBenchmark(b, entitiesNumber)
	view3 := NewView3[Pos, Vel, Acc](reg)
	subset := entities[:100]

	b.ResetTimer()
	for b.Loop() {
		view3Filter(view3, subset)
	}
}

// --- Helpery Pure ---

func view1PureAll(v *View) {
	for head := range PureAll1[Pos](v) {
		pos := head.V1
		pos.X += pos.X
	}
}

func view2PureAll(v *View) {
	for head := range PureAll2[Pos, Vel](v) {
		pos, vel := head.V1, head.V2
		pos.X += vel.X
	}
}

func view3PureAll(v *View) {
	for head := range PureAll3[Pos, Vel, Acc](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view4PureAll(v *View) {
	for head := range PureAll4[Pos, Vel, Acc, Mass](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view5PureAll(v *View) {
	for head := range PureAll5[Pos, Vel, Acc, Mass, Spin](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view6PureAll(v *View) {
	for head := range PureAll6[Pos, Vel, Acc, Mass, Spin, Char](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view7PureAll(v *View) {
	for head := range PureAll7[Pos, Vel, Acc, Mass, Spin, Char, Elec](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view8PureAll(v *View) {
	for head := range PureAll8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](v) {
		pos, vel, acc := head.V1, head.V2, head.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

// --- Benchmarki Pure All ---

func BenchmarkView1_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view1 := NewView1[Pos](reg)
	b.ResetTimer()
	for b.Loop() {
		view1PureAll(view1)
	}
}

func BenchmarkView2_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view2 := NewView2[Pos, Vel](reg)
	b.ResetTimer()
	for b.Loop() {
		view2PureAll(view2)
	}
}

func BenchmarkView3_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view3 := NewView3[Pos, Vel, Acc](reg)
	b.ResetTimer()
	for b.Loop() {
		view3PureAll(view3)
	}
}

func BenchmarkView4_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view4 := NewView4[Pos, Vel, Acc, Mass](reg)
	b.ResetTimer()
	for b.Loop() {
		view4PureAll(view4)
	}
}

func BenchmarkView5_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view5 := NewView5[Pos, Vel, Acc, Mass, Spin](reg)
	b.ResetTimer()
	for b.Loop() {
		view5PureAll(view5)
	}
}

func BenchmarkView6_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view6 := NewView6[Pos, Vel, Acc, Mass, Spin, Char](reg)
	b.ResetTimer()
	for b.Loop() {
		view6PureAll(view6)
	}
}

func BenchmarkView7_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view7 := NewView7[Pos, Vel, Acc, Mass, Spin, Char, Elec](reg)
	b.ResetTimer()
	for b.Loop() {
		view7PureAll(view7)
	}
}

func BenchmarkView8_PureAll(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view8 := NewView8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn](reg)
	b.ResetTimer()
	for b.Loop() {
		view8PureAll(view8)
	}
}

// --- Benchmarki Pure Filtered ---

func view3PureFilter(v *View, entities []Entity) {
	for head := range PureFilter3[Pos, Vel, Acc](v, entities) {
		pos, vel, acc := head.V1, head.V2, head.V3
		acc.X += vel.X
		pos.X += vel.X
	}
}

func BenchmarkView3_PureFiltered100(b *testing.B) {
	reg, entities := setupBenchmark(b, entitiesNumber)
	view3 := NewView3[Pos, Vel, Acc](reg)
	subset := entities[:100]

	b.ResetTimer()
	for b.Loop() {
		view3PureFilter(view3, subset)
	}
}
