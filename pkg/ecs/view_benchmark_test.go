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
type Mass struct{ M float32 }
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
		reg.AssignByID(e, massTypeId, unsafe.Pointer(&Mass{1}))
		reg.AssignByID(e, spinTypeId, unsafe.Pointer(&Spin{}))
		reg.AssignByID(e, charTypeId, unsafe.Pointer(&Char{}))
		reg.AssignByID(e, elecTypeId, unsafe.Pointer(&Elec{1}))
		reg.AssignByID(e, magnTypeId, unsafe.Pointer(&Magn{1}))
		entities = append(entities, e)
	}
	return reg, entities
}

func view1All(v *View1[Pos]) {
	for head := range All1(v) {
		pos := head.V1
		pos.X += pos.X
	}
}

func view2All(v *View2[Pos, Vel]) {
	for head := range All2(v) {
		pos, vel := head.V1, head.V2
		pos.X += vel.X
	}
}

func view3All(v *View3[Pos, Vel, Acc]) {
	for head, tail := range All3(v) {
		pos, vel, acc := head.V1, head.V2, tail.V3
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view4All(v *View4[Pos, Vel, Acc, Mass]) {
	for head, tail := range All4(v) {
		pos, vel, acc, mass := head.V1, head.V2, tail.V3, tail.V4
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view5All(v *View5[Pos, Vel, Acc, Mass, Spin]) {
	for head, tail := range All5(v) {
		pos, vel, acc, mass := head.V1, head.V2, tail.V3, tail.V4
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view6All(v *View6[Pos, Vel, Acc, Mass, Spin, Char]) {
	for head, tail := range All6(v) {
		pos, vel, acc, mass := head.V1, head.V2, head.V3, tail.V4
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view7All(v *View7[Pos, Vel, Acc, Mass, Spin, Char, Elec]) {
	for head, tail := range All7(v) {
		pos, vel, acc, mass := head.V1, head.V2, head.V3, tail.V4
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view8All(v *View8[Pos, Vel, Acc, Mass, Spin, Char, Elec, Magn]) {
	for head, tail := range All8(v) {
		pos, vel, acc, mass := head.V1, head.V2, head.V3, tail.V4
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func queryFiltered(v *View3[Pos, Vel, Acc], entities []Entity) {
	for head, tail := range Filter3(v, entities) {
		pos, vel, acc := head.V1, head.V2, tail.V3
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
		queryFiltered(view3, subset)
	}
}
