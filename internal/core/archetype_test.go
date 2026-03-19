package core

import (
	"reflect"
	"testing"
)

type TestStruct1 struct{ a int }

func TestInitArchetype(t *testing.T) {
	// given
	componentsRegistry := NewComponentsRegistry()
	archId := ArchetypeId(2)
	comp := componentsRegistry.GetOrRegister(reflect.TypeFor[TestStruct1]())
	mask := NewArchetypeMask(comp.ID)
	colsInfos := []ComponentInfo{comp}
	arch := Archetype{}

	// when
	arch.InitArchetype(archId, mask, colsInfos)

	// then
	if arch.Id != archId {
		t.Error("archetype Id is not set correctly")
	}
	if arch.Mask != mask {
		t.Error("archetype Mask is not set correctly")
	}
	if len(arch.Columns) != 2 {
		t.Error("archetype columns is not initialized correctly")
	}
	if arch.Columns[EntityColumnIndex].CompID != EntityID {
		t.Error("first column is not set correctly")
	}
	if len(arch.Memory.Pages) != 1 {
		t.Error("archetype memory pages is no initialized correctly")
	}
	if arch.Memory.Len != 0 {
		t.Error("archetype memory is not initialized correctly")
	}
}
