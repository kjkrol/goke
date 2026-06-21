package comp_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func newMetaIndex() comp.DefIndex {
	var c comp.DefIndex
	c.Init()
	return c
}

func TestMetaIndex_RegistrationAndMapping(t *testing.T) {
	c := newMetaIndex()

	t.Run("First registration", func(t *testing.T) {
		posType := reflect.TypeFor[position]()
		info := c.Intern(posType)

		if info.ID != 0 {
			t.Errorf("expected ID 0 for first registration, got %d", info.ID)
		}
		if info.Size != posType.Size() {
			t.Errorf("expected size %d, got %d", posType.Size(), info.Size)
		}
		if info.Type != posType {
			t.Errorf("expected type %v, got %v", posType, info.Type)
		}
	})

	t.Run("Idempotency (Register)", func(t *testing.T) {
		posType := reflect.TypeFor[position]()
		info1 := c.Intern(posType)
		info2 := c.Intern(posType)

		if info1.ID != info2.ID {
			t.Errorf("idempotency check failed: got IDs %d and %d", info1.ID, info2.ID)
		}
	})

	t.Run("Multiple types", func(t *testing.T) {
		fresh := newMetaIndex()

		id0 := fresh.Intern(reflect.TypeFor[position]()).ID
		id1 := fresh.Intern(reflect.TypeFor[velocity]()).ID
		id2 := fresh.Intern(reflect.TypeFor[rotation]()).ID

		if id0 != 0 || id1 != 1 || id2 != 2 {
			t.Errorf("incremental IDs failed: got %d, %d, %d", id0, id1, id2)
		}
	})
}

func TestMetaIndex_Lookup(t *testing.T) {
	c := newMetaIndex()
	posType := reflect.TypeFor[position]()
	info := c.Intern(posType)

	t.Run("Existing type", func(t *testing.T) {
		foundInfo, ok := c.ByType(posType)
		if !ok {
			t.Fatal("expected to find registered type")
		}
		if foundInfo.ID != info.ID {
			t.Errorf("lookup ID mismatch: expected %d, got %d", info.ID, foundInfo.ID)
		}
	})

	t.Run("Unregistered type", func(t *testing.T) {
		velType := reflect.TypeFor[velocity]()
		_, ok := c.ByType(velType)
		if ok {
			t.Error("expected ok=false for unregistered type")
		}
	})

	t.Run("Lookup by ID", func(t *testing.T) {
		foundInfo, ok := c.ByType(info.Type)
		if !ok {
			t.Fatal("expected to find info by type")
		}
		if foundInfo.Type != posType {
			t.Errorf("ID lookup type mismatch: expected %v, got %v", posType, foundInfo.Type)
		}
	})
}

func TestMetaIndex_GenericHelper(t *testing.T) {
	c := newMetaIndex()

	t.Run("Register via reflect.TypeFor matches manual reflection", func(t *testing.T) {
		genericInfo := c.Intern(reflect.TypeFor[position]())
		manualType := reflect.TypeFor[position]()
		manualInfo := c.Intern(manualType)

		if genericInfo.ID != manualInfo.ID || genericInfo.Type != manualInfo.Type {
			t.Error("generic helper does not match manual registration")
		}
	})
}

func TestMetaIndex_MetadataConsistency(t *testing.T) {
	c := newMetaIndex()

	t.Run("Size correctness", func(t *testing.T) {
		type complexStruct struct {
			a int64
			b float32
			c bool
		}
		cType := reflect.TypeFor[complexStruct]()
		info := c.Intern(cType)

		if info.Size != cType.Size() {
			t.Errorf("size mismatch: expected %d, got %d", cType.Size(), info.Size)
		}
	})

	t.Run("Pointers vs Values", func(t *testing.T) {
		valType := reflect.TypeFor[position]()
		ptrType := reflect.TypeOf(&position{})

		infoVal := c.Intern(valType)
		infoPtr := c.Intern(ptrType)

		if infoVal.ID == infoPtr.ID {
			t.Error("value and pointer types must have distinct component IDs")
		}
	})
}

func TestMetaIndex_EdgeCases(t *testing.T) {
	c := newMetaIndex()

	t.Run("Empty structures", func(t *testing.T) {
		type empty struct{}
		info := c.Intern(reflect.TypeFor[empty]())

		if info.Size != 0 {
			t.Errorf("expected size 0 for empty struct, got %d", info.Size)
		}
	})

	t.Run("Primitive types", func(t *testing.T) {
		intInfo := c.Intern(reflect.TypeFor[int]())
		strInfo := c.Intern(reflect.TypeFor[string]())

		if intInfo.ID == strInfo.ID {
			t.Error("primitive types must have unique IDs")
		}
		if intInfo.Size != reflect.TypeFor[int]().Size() {
			t.Error("primitive type size mismatch")
		}
	})
}
