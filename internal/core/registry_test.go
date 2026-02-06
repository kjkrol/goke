package core

import (
	"reflect"
	"testing"
)

func TestViewReactivity(t *testing.T) {
	// 1. Setup: Create Registry and define component types
	reg := NewRegistry(DefaultRegistryConfig())

	type Position struct{ X, Y float32 }
	type TagA struct{}

	posTypeInfo := reg.RegisterComponentType(reflect.TypeFor[Position]())
	tagTypeInfo := reg.RegisterComponentType(reflect.TypeFor[TagA]())

	// 2. Initialize View BEFORE creating any matching entities
	// This view looks for entities having both Position AND TagA
	blueprint := NewBlueprint(reg)
	blueprint.WithComp(posTypeInfo)
	blueprint.WithTag(tagTypeInfo.ID)
	view := NewView(blueprint, reg)

	// Initial check: The view should be empty because no archetypes exist yet
	if len(view.Baked) != 0 {
		t.Errorf("Expected 0 baked archetypes initially, got %d", len(view.Baked))
	}

	// 3. Action: Create an entity that forces a new matching archetype
	e1 := reg.CreateEntity()

	// Assigning components triggers ArchetypeRegistry.getOrRegister
	// which should notify the ViewRegistry
	posPtr, _ := reg.AllocateByID(e1, posTypeInfo)
	*(*Position)(posPtr) = Position{X: 10, Y: 20}
	_, _ = reg.AllocateByID(e1, tagTypeInfo)

	// 4. Verification: Did the View receive the new archetype?
	if len(view.Baked) == 0 {
		t.Fatal("View failed to detect new archetype created after view initialization!")
	}

	// Verify the data inside baked reflects the added entity
	foundMatchingArch := false
	for _, b := range view.Baked {
		if *b.Len == 1 {
			foundMatchingArch = true
			break
		}
	}

	if !foundMatchingArch {
		t.Error("View has the archetype, but the entity count (len) is incorrect")
	}

	// 5. Negative Test: Create an entity that DOES NOT match the mask
	// (Has Position, but lacks TagA)
	beforeCount := len(view.Baked)
	e2 := reg.CreateEntity()
	posPtr2, _ := reg.AllocateByID(e2, posTypeInfo)
	*(*Position)(posPtr2) = Position{X: 30, Y: 40}

	if len(view.Baked) != beforeCount {
		t.Errorf("View incorrectly added an archetype that does not satisfy the TagA requirement")
	}
}
