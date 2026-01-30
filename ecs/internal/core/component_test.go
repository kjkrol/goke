package core_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/ecs/internal/core"
)

// 1. Registration and Type Mapping
func TestComponentsRegistry_RegistrationAndMapping(t *testing.T) {
	registry := core.NewComponentsRegistry()

	// Case: First registration
	// Checking if a new type is correctly registered with ID 0, correct size and type.
	t.Run("First registration", func(t *testing.T) {
		posType := reflect.TypeFor[position]()
		info := registry.GetOrRegister(posType)

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

	// Case: Idempotency (GetOrRegister)
	// Verifying that calling GetOrRegister again for the same type returns exactly the same data.
	t.Run("Idempotency (GetOrRegister)", func(t *testing.T) {
		posType := reflect.TypeFor[position]()
		info1 := registry.GetOrRegister(posType)
		info2 := registry.GetOrRegister(posType)

		if info1.ID != info2.ID {
			t.Errorf("idempotency check failed: got IDs %d and %d", info1.ID, info2.ID)
		}
	})

	// Case: Multiple types
	// Registering several different structures and ensuring each receives a unique, incremental ID.
	t.Run("Multiple types", func(t *testing.T) {
		registry := core.NewComponentsRegistry() // Fresh registry for clean IDs

		id0 := registry.GetOrRegister(reflect.TypeFor[position]()).ID
		id1 := registry.GetOrRegister(reflect.TypeFor[velocity]()).ID
		id2 := registry.GetOrRegister(reflect.TypeFor[rotation]()).ID

		if id0 != 0 || id1 != 1 || id2 != 2 {
			t.Errorf("incremental IDs failed: got %d, %d, %d", id0, id1, id2)
		}
	})
}

// 2. Information Lookup
func TestComponentsRegistry_Lookup(t *testing.T) {
	registry := core.NewComponentsRegistry()
	posType := reflect.TypeFor[position]()
	info := registry.GetOrRegister(posType)

	// Case: Existing type
	// Checking if the Get method correctly finds a previously registered component.
	t.Run("Existing type", func(t *testing.T) {
		foundInfo, ok := registry.Get(posType)
		if !ok {
			t.Fatal("expected to find registered type")
		}
		if foundInfo.ID != info.ID {
			t.Errorf("lookup ID mismatch: expected %d, got %d", info.ID, foundInfo.ID)
		}
	})

	// Case: Unregistered type
	// Verifying that Get returns false for a type that has never appeared in the registry.
	t.Run("Unregistered type", func(t *testing.T) {
		velType := reflect.TypeFor[velocity]()
		_, ok := registry.Get(velType)
		if ok {
			t.Error("expected ok=false for unregistered type")
		}
	})

	// Case: Lookup by ID
	// Checking if ComponentInfo can be found using only its ComponentID.
	t.Run("Lookup by ID", func(t *testing.T) {
		foundInfo, ok := registry.Get(info.Type)
		if !ok {
			t.Fatal("expected to find info in idToInfo map")
		}
		if foundInfo.Type != posType {
			t.Errorf("ID lookup type mismatch: expected %v, got %v", posType, foundInfo.Type)
		}
	})
}

// 3. Generic Helper
func TestComponentsRegistry_GenericHelper(t *testing.T) {
	registry := core.NewComponentsRegistry()

	// Case: ensureComponentRegistered
	// Testing if the generic call ensureComponentRegistered[T](reg) works identically to manual reflection.
	t.Run("ensureComponentRegistered helper", func(t *testing.T) {
		genericInfo := registry.GetOrRegister(reflect.TypeFor[position]())
		manualType := reflect.TypeFor[position]()
		manualInfo := registry.GetOrRegister(manualType)

		if genericInfo.ID != manualInfo.ID || genericInfo.Type != manualInfo.Type {
			t.Error("generic helper does not match manual registration")
		}
	})
}

// 4. Metadata Consistency
func TestComponentsRegistry_MetadataConsistency(t *testing.T) {
	registry := core.NewComponentsRegistry()

	// Case: Size correctness
	// Verifying that Size in ComponentInfo actually corresponds to the structure size.
	t.Run("Size correctness", func(t *testing.T) {
		type complexStruct struct {
			a int64
			b float32
			c bool
		}
		cType := reflect.TypeFor[complexStruct]()
		info := registry.GetOrRegister(cType)

		if info.Size != cType.Size() {
			t.Errorf("size mismatch: expected %d, got %d", cType.Size(), info.Size)
		}
	})

	// Case: Pointers vs Values
	// Checking if Position and *Position are treated as two distinct components.
	t.Run("Pointers vs Values", func(t *testing.T) {
		valType := reflect.TypeFor[position]()
		ptrType := reflect.TypeOf(&position{})

		infoVal := registry.GetOrRegister(valType)
		infoPtr := registry.GetOrRegister(ptrType)

		if infoVal.ID == infoPtr.ID {
			t.Error("registry should treat value and pointer types as distinct components")
		}
	})
}

// 5. Edge Cases
func TestComponentsRegistry_EdgeCases(t *testing.T) {
	registry := core.NewComponentsRegistry()

	// Case: Empty structures
	// Checking how the registry handles struct{} (should register with size 0).
	t.Run("Empty structures", func(t *testing.T) {
		type empty struct{}
		info := registry.GetOrRegister(reflect.TypeFor[empty]())

		if info.Size != 0 {
			t.Errorf("expected size 0 for empty struct, got %d", info.Size)
		}
	})

	// Case: Primitive types
	// Registering built-in types like int or string as components.
	t.Run("Primitive types", func(t *testing.T) {
		intInfo := registry.GetOrRegister(reflect.TypeFor[int]())
		strInfo := registry.GetOrRegister(reflect.TypeFor[string]())

		if intInfo.ID == strInfo.ID {
			t.Error("primitive types must have unique IDs")
		}
		if intInfo.Size != reflect.TypeFor[int]().Size() {
			t.Error("primitive type size mismatch")
		}
	})
}
