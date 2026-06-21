package arch

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

type testStruct1 struct{ a int }

func TestInitArchetype(t *testing.T) {
	mi := newCatalog()
	archId := ID(2)
	compDef := mi.Intern(reflect.TypeFor[testStruct1]())
	set := comp.Composition{}.With(compDef)
	a := Archetype{}

	a.Init(archId, set)

	if a.Id != archId {
		t.Error("archetype Id is not set correctly")
	}
	if a.Mask() != set.Mask {
		t.Error("archetype Mask is not set correctly")
	}
	if a.Table.Len() != 0 {
		t.Error("archetype memory is not initialized correctly")
	}
}
