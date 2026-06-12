package reg

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/core"
)

func TestViewReactivity(t *testing.T) {
	registry := NewRegistry(DefaultRegistryConfig())

	type Position struct{ X, Y float32 }
	type TagA struct{}

	posTypeInfo := registry.RegisterComponentType(reflect.TypeFor[Position]())
	tagTypeInfo := registry.RegisterComponentType(reflect.TypeFor[TagA]())

	blueprint := NewBlueprint(registry)
	blueprint.WithComp(posTypeInfo)
	blueprint.WithTag(tagTypeInfo.ID)
	view := NewView(blueprint, []core.ComponentInfo{posTypeInfo}, registry)

	if len(view.Baked) != 0 {
		t.Errorf("Expected 0 baked archetypes initially, got %d", len(view.Baked))
	}

	e1 := registry.CreateEntity()

	posPtr, _ := registry.AllocateByID(e1, posTypeInfo)
	*(*Position)(posPtr) = Position{X: 10, Y: 20}
	_, _ = registry.AllocateByID(e1, tagTypeInfo)

	if len(view.Baked) == 0 {
		t.Fatal("View failed to detect new archetype created after view initialization!")
	}

	foundMatchingArch := false
	for _, b := range view.Baked {
		if b.Arch.Len() == 1 {
			foundMatchingArch = true
			break
		}
	}

	if !foundMatchingArch {
		t.Error("View has the archetype, but the entity count (len) is incorrect")
	}

	beforeCount := len(view.Baked)
	e2 := registry.CreateEntity()
	posPtr2, _ := registry.AllocateByID(e2, posTypeInfo)
	*(*Position)(posPtr2) = Position{X: 30, Y: 40}

	if len(view.Baked) != beforeCount {
		t.Errorf("View incorrectly added an archetype that does not satisfy the TagA requirement")
	}
}
