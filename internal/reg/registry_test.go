package reg

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/query"
)

func TestViewReactivity(t *testing.T) {
	var registry Registry
	registry.Init(DefaultConfig())

	type Position struct{ X, Y float32 }
	type TagA struct{}

	posTypeInfo := registry.RegCompType(reflect.TypeFor[Position]())
	tagTypeInfo := registry.RegCompType(reflect.TypeFor[TagA]())

	blueprint := comp.NewBlueprint()
	blueprint.Comp(posTypeInfo)
	blueprint.Tag(tagTypeInfo.ID)
	query := query.NewView(blueprint, []comp.Meta{posTypeInfo}, &registry.ArchCatalog, &registry.ViewRegistry)

	if len(query.MatchedArchs) != 0 {
		t.Errorf("Expected 0 baked archetypes initially, got %d", len(query.MatchedArchs))
	}

	e1 := registry.CreateEntity()

	posPtr, _ := registry.UpsertComp(e1, posTypeInfo)
	*(*Position)(posPtr) = Position{X: 10, Y: 20}
	_, _ = registry.UpsertComp(e1, tagTypeInfo)

	if len(query.MatchedArchs) == 0 {
		t.Fatal("View failed to detect new archetype created after query initialization!")
	}

	foundMatchingArch := false
	for _, b := range query.MatchedArchs {
		if b.Table.Len == 1 {
			foundMatchingArch = true
			break
		}
	}

	if !foundMatchingArch {
		t.Error("View has the archetype, but the entity count (len) is incorrect")
	}

	beforeCount := len(query.MatchedArchs)
	e2 := registry.CreateEntity()
	posPtr2, _ := registry.UpsertComp(e2, posTypeInfo)
	*(*Position)(posPtr2) = Position{X: 30, Y: 40}

	if len(query.MatchedArchs) != beforeCount {
		t.Errorf("View incorrectly added an archetype that does not satisfy the TagA requirement")
	}
}
