package arch

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

type testStruct1 struct{ a int }

func TestInitArchetype(t *testing.T) {
	mi := newDefIndex()
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
	if a.Len() != 0 {
		t.Error("Archetype.Len should mirror the table's row count")
	}
}

func TestArchetype_Reset(t *testing.T) {
	mi := newDefIndex()
	compDef := mi.Intern(reflect.TypeFor[testStruct1]())
	set := comp.Composition{}.With(compDef)
	a := Archetype{}
	a.Init(ID(3), set)

	a.Reset()

	if a.Id != NullID {
		t.Errorf("expected Id to be reset to NullID, got %d", a.Id)
	}
	if !a.Mask().IsEmpty() {
		t.Error("expected Mask to be cleared after Reset")
	}
}
