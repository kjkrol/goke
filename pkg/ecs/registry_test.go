package ecs

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestViewReactivity(t *testing.T) {
	// 1. Setup: Create Registry and define component types
	reg := NewRegistry()

	type Position struct{ X, Y float32 }
	type TagA struct{}

	posTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Position]())
	tagTypeInfo := reg.RegisterComponentType(reflect.TypeFor[TagA]())

	// 2. Initialize View BEFORE creating any matching entities
	// This view looks for entities having both Position AND TagA
	view := NewViewBuilder(reg).
		OnType(posTypeInfo.ID).
		OnTag(tagTypeInfo.ID).
		Build()

	// Initial check: The view should be empty because no archetypes exist yet
	if len(view.baked) != 0 {
		t.Errorf("Expected 0 baked archetypes initially, got %d", len(view.baked))
	}

	// 3. Action: Create an entity that forces a new matching archetype
	e1 := reg.CreateEntity()

	// Assigning components triggers ArchetypeRegistry.getOrRegister
	// which should notify the ViewRegistry
	reg.AssignByID(e1, posTypeInfo.ID, unsafe.Pointer(&Position{10, 20}))
	reg.AssignByID(e1, tagTypeInfo.ID, unsafe.Pointer(&TagA{}))

	// 4. Verification: Did the View receive the new archetype?
	if len(view.baked) == 0 {
		t.Fatal("View failed to detect new archetype created after view initialization!")
	}

	// Verify the data inside baked reflects the added entity
	foundMatchingArch := false
	for _, b := range view.baked {
		if *b.len == 1 {
			foundMatchingArch = true
			break
		}
	}

	if !foundMatchingArch {
		t.Error("View has the archetype, but the entity count (len) is incorrect")
	}

	// 5. Negative Test: Create an entity that DOES NOT match the mask
	// (Has Position, but lacks TagA)
	beforeCount := len(view.baked)
	e2 := reg.CreateEntity()
	reg.AssignByID(e2, posTypeInfo.ID, unsafe.Pointer(&Position{30, 40}))

	if len(view.baked) != beforeCount {
		t.Errorf("View incorrectly added an archetype that does not satisfy the TagA requirement")
	}
}
