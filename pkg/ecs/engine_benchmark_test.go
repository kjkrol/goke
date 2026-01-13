package ecs

import (
	"testing"
)

const entitiesNumber = 1000

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }

func setupBenchmark(_ *testing.B, count int) (*Registry, []Entity) {
	reg := newRegistry()
	var entities []Entity
	for range count {
		e := reg.entitiesRegistry.create()
		assign(reg, e, Pos{1, 1})
		assign(reg, e, Vel{1, 1})
		assign(reg, e, Acc{1, 1})
		entities = append(entities, e)
	}
	return reg, entities
}

func query1All(v *View1[Pos]) {
	for _, row := range v.All() {
		pos := row.Values()
		pos.X += pos.X
	}
}

func query3All(v *View3[Pos, Vel, Acc]) {
	for _, row := range v.All() {
		pos, vel, acc := row.Values()
		pos.X += vel.X + acc.X
	}
}

func queryFiltered(v *View3[Pos, Vel, Acc], entities []Entity) {
	for _, row := range v.Filtered(entities) {
		row.V1.X += row.V2.X + row.V3.X
	}
}

func BenchmarkView1_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view1 := NewView1[Pos](reg)

	b.ResetTimer()
	for b.Loop() {
		query1All(view1)
	}
}

func BenchmarkView3_All(b *testing.B) {
	reg, _ := setupBenchmark(b, entitiesNumber)
	view3 := NewView3[Pos, Vel, Acc](reg)

	b.ResetTimer()
	for b.Loop() {
		query3All(view3)
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
