package query

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
)

func newQueryCatalog() (*Catalog, *comp.Catalog, *ent.Manager) {
	var cc comp.Catalog
	cc.Init()
	cat := new(Catalog)
	var em ent.Manager
	cat.Init(&cc, &em, DefaultConfig())
	em.Init(ent.DefaultConfig(), cat.OnArchetypeCreated)
	return cat, &cc, &em
}

func TestViewReactivity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	type Position struct{ X, Y float32 }
	type TagA struct{}

	posTypeInfo := cc.Intern(reflect.TypeFor[Position]())
	tagTypeInfo := cc.Intern(reflect.TypeFor[TagA]())

	blueprint := comp.NewBlueprint()
	blueprint.Comp(posTypeInfo)
	blueprint.Tag(tagTypeInfo.ID)
	v := cat.AddView(blueprint)

	if len(v.BakedTables) != 0 {
		t.Errorf("Expected 0 baked tables initially, got %d", len(v.BakedTables))
	}

	e1 := em.Create()

	posPtr, _ := em.UpsertComp(e1, posTypeInfo)
	*(*Position)(posPtr) = Position{X: 10, Y: 20}
	_, _ = em.UpsertComp(e1, tagTypeInfo)

	if len(v.BakedTables) == 0 {
		t.Fatal("View failed to detect new archetype created after query initialization!")
	}

	foundMatchingArch := false
	for _, b := range v.BakedTables {
		if b.Table.Len == 1 {
			foundMatchingArch = true
			break
		}
	}
	if !foundMatchingArch {
		t.Error("View has the archetype, but the entity count (len) is incorrect")
	}

	beforeCount := len(v.BakedTables)
	e2 := em.Create()
	posPtr2, _ := em.UpsertComp(e2, posTypeInfo)
	*(*Position)(posPtr2) = Position{X: 30, Y: 40}

	if len(v.BakedTables) != beforeCount {
		t.Errorf("View incorrectly added an archetype that does not satisfy the TagA requirement")
	}
}
