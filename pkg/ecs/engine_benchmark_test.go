package ecs

import (
	"testing"
)

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

func queryAll(q *View3[Pos, Vel, Acc]) {
	for _, row := range q.All() {
		row.V1.X += row.V2.X + row.V3.X
	}
}

//func queryFiltered(q *View3[Pos, Vel, Acc], entities []Entity) {
//	for _, row := range q.Filtered(entities) {
//		row.V1.X += row.V2.X + row.V3.X
//	}
//}

func BenchmarkView3_All(b *testing.B) {
	reg, _ := setupBenchmark(b, 10000)
	q := NewView3[Pos, Vel, Acc](reg)

	b.ResetTimer()
	for b.Loop() {
		queryAll(q)
	}
}

//func BenchmarkView3_Filtered(b *testing.B) {
//	reg, entities := setupBenchmark(b, 10000)
//	q := NewView3[Pos, Vel, Acc](reg)
//	subset := entities[:100]
//
//	b.ResetTimer()
//	for b.Loop() {
//		queryFiltered(q, subset)
//	}
//}
