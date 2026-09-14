package chunk

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/comp"
)

type benchPOD struct{ X, Y float32 }

// BenchmarkPack_AddChunks isolates chunk allocation from the entity-creation
// harness.
func BenchmarkPack_AddChunks(b *testing.B) {
	var l Layout
	l.Init([]comp.Def{
		{ID: 1, Size: unsafe.Sizeof(benchPOD{}), Align: 4, Type: reflect.TypeFor[benchPOD]()},
		{ID: 2, Size: 8, Align: 8, Type: reflect.TypeFor[float64]()},
		{ID: 3, Size: 4, Align: 4, Type: reflect.TypeFor[float32]()},
		{ID: 4, Size: 8, Align: 8, Type: reflect.TypeFor[int64]()},
	})
	b.ReportAllocs()
	for b.Loop() {
		var g Pack
		g.Init(l)
		g.AddChunks(4)
	}
}
