package goke_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

func TestOptColumn_AtReadsTheEntitysOwnValue(t *testing.T) {
	ecs := goke.New()
	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	posCol := goke.NewColumn(reflect.TypeFor[Position]())
	velCol := goke.NewOptColumn(reflect.TypeFor[Velocity]())
	var query *goke.Query
	want := map[uid.UID64]Velocity{}
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&pos, &vel)
		f.Create(3)
		for f.Next() {
			for i, id := range f.IDs {
				v := Velocity{VX: float32(len(want) + 1), VY: float32(i)}
				vel.Slice(&f.Cursor)[i] = v
				want[id] = v
			}
		}
		query = si.NewQueryBuilder(posCol).Optional(velCol).Build()
	}})

	// Slot 0 hides a wrong sign, so every entity of the chunk is read.
	for id, w := range want {
		if !query.Seek(id) {
			t.Fatalf("Seek(%v) failed", id)
		}
		got := velCol.At(query.Cursor())
		if got == nil {
			t.Fatalf("entity %v: the present optional column has no value", id)
		}
		if v := *(*Velocity)(unsafe.Pointer(got)); v != w {
			t.Errorf("entity %v read %v, want %v", id, v, w)
		}
	}
}
