package arch

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

type testStruct1 struct{ a int }

func TestInitArchetype(t *testing.T) {
	mi := comp.NewCatalog()
	archId := ID(2)
	compMeta := mi.Register(reflect.TypeFor[testStruct1]())
	set := comp.Composition{}.With(compMeta)
	a := Archetype{}

	a.Init(archId, set)

	if a.Id != archId {
		t.Error("archetype Id is not set correctly")
	}
	if a.Mask() != set.Mask {
		t.Error("archetype Mask is not set correctly")
	}
	if a.Table.NumColumns() != 2 {
		t.Error("archetype columns is not initialized correctly")
	}
	if a.Table.GetEntityColumn().CompID != comp.EntityID {
		t.Error("first column is not set correctly")
	}
	if len(a.Table.Chunks) != 1 {
		t.Error("archetype memory pages is no initialized correctly")
	}
	if a.Table.Len != 0 {
		t.Error("archetype memory is not initialized correctly")
	}
}
