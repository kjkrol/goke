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

func setupBenchmark(_ *testing.B, count int) (*Registry, []Entity) {
	reg := NewRegistry()
	posTypeId := reg.RegisterComponentType(reflect.TypeFor[Pos]())
	velTypeId := reg.RegisterComponentType(reflect.TypeFor[Vel]())
	accTypeId := reg.RegisterComponentType(reflect.TypeFor[Acc]())
	massTypeId := reg.RegisterComponentType(reflect.TypeFor[Mass]())
	spinTypeId := reg.RegisterComponentType(reflect.TypeFor[Spin]())
	charTypeId := reg.RegisterComponentType(reflect.TypeFor[Char]())

	var entities []Entity
	for range count {
		e := reg.CreateEntity()
		reg.AssignByID(e, posTypeId, unsafe.Pointer(&Pos{1, 1}))
		reg.AssignByID(e, velTypeId, unsafe.Pointer(&Vel{1, 1}))
		reg.AssignByID(e, accTypeId, unsafe.Pointer(&Acc{1, 1}))
		reg.AssignByID(e, massTypeId, unsafe.Pointer(&Mass{1}))
		reg.AssignByID(e, spinTypeId, unsafe.Pointer(&Spin{}))
		reg.AssignByID(e, charTypeId, unsafe.Pointer(&Char{}))
		entities = append(entities, e)
	}
	return reg, entities
}

func view1All(v *View1[Pos]) {
	for _, row := range v.All() {
		pos := row.Values()
		pos.X += pos.X
	}
}

func view2All(v *View2[Pos, Vel]) {
	for _, row := range v.All() {
		pos, vel := row.Values()
		pos.X += vel.X
	}
}

func view3All(v *View3[Pos, Vel, Acc]) {
	for _, row := range v.All() {
		pos, vel, acc := row.Values()
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view4All(v *View4[Pos, Vel, Acc, Mass]) {
	for _, row := range v.All() {
		pos, vel, acc, mass := row.Values()
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view5All(v *View5[Pos, Vel, Acc, Mass, Spin]) {
	for _, row := range v.All() {
		pos, vel, acc, mass, _ := row.Values()
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func view6All(v *View6[Pos, Vel, Acc, Mass, Spin, Char]) {
	for _, row := range v.All() {
		pos, vel, acc, mass, _, _ := row.Values()
		acc.X += mass.M
		vel.X += acc.X
		pos.X += vel.X
	}
}

func queryFiltered(v *View3[Pos, Vel, Acc], entities []Entity) {
	for _, row := range v.Filtered(entities) {
		pos, vel, acc := row.Values()
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

func BenchmarkView3_Filtered100(b *testing.B) {
	reg, entities := setupBenchmark(b, entitiesNumber)
	view3 := NewView3[Pos, Vel, Acc](reg)
	subset := entities[:100]

	b.ResetTimer()
	for b.Loop() {
		queryFiltered(view3, subset)
	}
}
