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
	for i := 0; i < count; i++ {
		e := reg.createEntity()
		assign(reg, e, Pos{1, 1})
		assign(reg, e, Vel{1, 1})
		assign(reg, e, Acc{1, 1})
		entities = append(entities, e)
	}
	return reg, entities
}

// Benchmark wersji Composed (z mapperem i matchAndIterate)
func BenchmarkQuery3_Composed(b *testing.B) {
	reg, _ := setupBenchmark(b, 10000)
	q := NewQuery3[Pos, Vel, Acc](reg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, row := range q.All() {
			row.V1.X += row.V2.X + row.V3.X
		}
	}
}

// Benchmark wersji Filtered (tylko 100 encji z 10000)
func BenchmarkQuery3_Filtered(b *testing.B) {
	reg, entities := setupBenchmark(b, 10000)
	q := NewQuery3[Pos, Vel, Acc](reg)
	subset := entities[:100] // Tylko 1% encji

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, row := range q.Filtered(subset) {
			row.V1.X += row.V2.X + row.V3.X
		}
	}
}
