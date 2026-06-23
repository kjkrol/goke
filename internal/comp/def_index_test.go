package comp_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func newDefIndex() comp.DefIndex {
	var c comp.DefIndex
	c.Init()
	return c
}

func TestDefIndex_RegistrationAndMapping(t *testing.T) {
	c := newDefIndex()

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
		fresh := newDefIndex()

		id0 := fresh.Intern(reflect.TypeFor[position]()).ID
		id1 := fresh.Intern(reflect.TypeFor[velocity]()).ID
		id2 := fresh.Intern(reflect.TypeFor[rotation]()).ID

		if id0 != 0 || id1 != 1 || id2 != 2 {
			t.Errorf("incremental IDs failed: got %d, %d, %d", id0, id1, id2)
		}
	})
}

func TestDefIndex_Lookup(t *testing.T) {
	c := newDefIndex()
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

func TestDefIndex_GenericHelper(t *testing.T) {
	c := newDefIndex()

	t.Run("Register via reflect.TypeFor matches manual reflection", func(t *testing.T) {
		genericInfo := c.Intern(reflect.TypeFor[position]())
		manualType := reflect.TypeFor[position]()
		manualInfo := c.Intern(manualType)

		if genericInfo.ID != manualInfo.ID || genericInfo.Type != manualInfo.Type {
			t.Error("generic helper does not match manual registration")
		}
	})
}

func TestDefIndex_MetadataConsistency(t *testing.T) {
	c := newDefIndex()

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

func TestDefIndex_EdgeCases(t *testing.T) {
	c := newDefIndex()

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

func TestDefIndex_ByID(t *testing.T) {
	c := newDefIndex()
	posType := reflect.TypeFor[position]()
	info := c.Intern(posType)

	if got := c.ByID(info.ID); got.Type != posType {
		t.Errorf("expected ByID(%d) to return the interned type, got %v", info.ID, got.Type)
	}

	if got := c.ByID(comp.ID(99)); got.Type != nil {
		t.Errorf("expected ByID for an unregistered ID to return a zero Def, got %+v", got)
	}
}

func TestDefIndex_Reset(t *testing.T) {
	c := newDefIndex()
	posType := reflect.TypeFor[position]()
	info := c.Intern(posType)

	c.Reset()

	if _, ok := c.ByType(posType); ok {
		t.Error("expected ByType to find nothing after Reset")
	}
	if got := c.ByID(info.ID); got.Type != nil {
		t.Errorf("expected ByID to return a zero Def after Reset, got %+v", got)
	}

	// Registration must restart from ID 0 after Reset.
	again := c.Intern(posType)
	if again.ID != 0 {
		t.Errorf("expected first registration after Reset to get ID 0, got %d", again.ID)
	}
}

func TestDefIndex_ResetOnZeroValue(t *testing.T) {
	var c comp.DefIndex // never Init'd — typeIndex is nil

	c.Reset()

	posType := reflect.TypeFor[position]()
	info := c.Intern(posType)
	if info.ID != 0 {
		t.Errorf("expected first registration after Reset to get ID 0, got %d", info.ID)
	}
}

func TestDefIndex_InternPanicsWhenFull(t *testing.T) {
	c := newDefIndex()
	for i := 1; i <= comp.MaxComponents; i++ {
		c.Intern(reflect.ArrayOf(i, reflect.TypeFor[byte]()))
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when exceeding MaxComponents")
		}
	}()
	c.Intern(reflect.ArrayOf(comp.MaxComponents+1, reflect.TypeFor[byte]()))
}
