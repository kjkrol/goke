package comp_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func TestCatalog_RegistrationAndMapping(t *testing.T) {
	c := comp.NewCatalog()

	t.Run("First registration", func(t *testing.T) {
		posType := reflect.TypeFor[position]()
		info := c.Register(posType)

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
		info1 := c.Register(posType)
		info2 := c.Register(posType)

		if info1.ID != info2.ID {
			t.Errorf("idempotency check failed: got IDs %d and %d", info1.ID, info2.ID)
		}
	})

	t.Run("Multiple types", func(t *testing.T) {
		fresh := comp.NewCatalog()

		id0 := fresh.Register(reflect.TypeFor[position]()).ID
		id1 := fresh.Register(reflect.TypeFor[velocity]()).ID
		id2 := fresh.Register(reflect.TypeFor[rotation]()).ID

		if id0 != 0 || id1 != 1 || id2 != 2 {
			t.Errorf("incremental IDs failed: got %d, %d, %d", id0, id1, id2)
		}
	})
}

func TestCatalog_Lookup(t *testing.T) {
	c := comp.NewCatalog()
	posType := reflect.TypeFor[position]()
	info := c.Register(posType)

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

func TestCatalog_GenericHelper(t *testing.T) {
	c := comp.NewCatalog()

	t.Run("Register via reflect.TypeFor matches manual reflection", func(t *testing.T) {
		genericInfo := c.Register(reflect.TypeFor[position]())
		manualType := reflect.TypeFor[position]()
		manualInfo := c.Register(manualType)

		if genericInfo.ID != manualInfo.ID || genericInfo.Type != manualInfo.Type {
			t.Error("generic helper does not match manual registration")
		}
	})
}

func TestCatalog_MetadataConsistency(t *testing.T) {
	c := comp.NewCatalog()

	t.Run("Size correctness", func(t *testing.T) {
		type complexStruct struct {
			a int64
			b float32
			c bool
		}
		cType := reflect.TypeFor[complexStruct]()
		info := c.Register(cType)

		if info.Size != cType.Size() {
			t.Errorf("size mismatch: expected %d, got %d", cType.Size(), info.Size)
		}
	})

	t.Run("Pointers vs Values", func(t *testing.T) {
		valType := reflect.TypeFor[position]()
		ptrType := reflect.TypeOf(&position{})

		infoVal := c.Register(valType)
		infoPtr := c.Register(ptrType)

		if infoVal.ID == infoPtr.ID {
			t.Error("value and pointer types must have distinct component IDs")
		}
	})
}

func TestCatalog_EdgeCases(t *testing.T) {
	c := comp.NewCatalog()

	t.Run("Empty structures", func(t *testing.T) {
		type empty struct{}
		info := c.Register(reflect.TypeFor[empty]())

		if info.Size != 0 {
			t.Errorf("expected size 0 for empty struct, got %d", info.Size)
		}
	})

	t.Run("Primitive types", func(t *testing.T) {
		intInfo := c.Register(reflect.TypeFor[int]())
		strInfo := c.Register(reflect.TypeFor[string]())

		if intInfo.ID == strInfo.ID {
			t.Error("primitive types must have unique IDs")
		}
		if intInfo.Size != reflect.TypeFor[int]().Size() {
			t.Error("primitive type size mismatch")
		}
	})
}
