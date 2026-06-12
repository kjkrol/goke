package arch

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/internal/mem"
)

type testStruct1 struct{ a int }

func TestInitArchetype(t *testing.T) {
	componentsRegistry := core.NewComponentsRegistry()
	archId := core.ArchetypeId(2)
	comp := componentsRegistry.GetOrRegister(reflect.TypeFor[testStruct1]())
	mask := core.NewArchetypeMask(comp.ID)
	colsInfos := []core.ComponentInfo{comp}
	arch := Archetype{}

	arch.InitArchetype(archId, mask, colsInfos)

	if arch.Id != archId {
		t.Error("archetype Id is not set correctly")
	}
	if arch.Mask != mask {
		t.Error("archetype Mask is not set correctly")
	}
	if len(arch.Columns) != 2 {
		t.Error("archetype columns is not initialized correctly")
	}
	if arch.Columns[mem.EntityColumnIndex].CompID != core.EntityID {
		t.Error("first column is not set correctly")
	}
	if len(arch.Memory.Pages) != 1 {
		t.Error("archetype memory pages is no initialized correctly")
	}
	if arch.Memory.Len != 0 {
		t.Error("archetype memory is not initialized correctly")
	}
}
