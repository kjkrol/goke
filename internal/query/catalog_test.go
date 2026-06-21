package query

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
)

func newQueryCatalog() (*Catalog, *comp.DefIndex, *ent.Manager) {
	var cc comp.DefIndex
	cc.Init()
	cat := new(Catalog)
	var em ent.Manager
	cat.Init(&cc, &em.AddressBook.Index, &em.ArchCatalog, DefaultConfig())
	em.Init(ent.DefaultConfig(), cat.OnArchetypeCreated)
	return cat, &cc, &em
}

func TestViewReactivity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	type Position struct{ X, Y float32 }
	type TagA struct{}

	posTypeInfo := cc.Intern(reflect.TypeFor[Position]())
	tagTypeInfo := cc.Intern(reflect.TypeFor[TagA]())

	var blueprint comp.Blueprint
	blueprint.Comp(posTypeInfo)
	blueprint.Tag(tagTypeInfo.ID)
	v := cat.AddView(&blueprint)

	if len(v.BakedTables) != 0 {
		t.Errorf("Expected 0 baked tables initially, got %d", len(v.BakedTables))
	}

	var b1 comp.Blueprint
	b1.Comp(posTypeInfo)
	b1.Tag(tagTypeInfo.ID)
	f1 := em.CreateFactory(b1)
	f1.Create(1)
	f1.Next()

	if len(v.BakedTables) == 0 {
		t.Fatal("View failed to detect new archetype created after query initialization!")
	}

	foundMatchingArch := false
	for _, b := range v.BakedTables {
		if b.Table.Len() == 1 {
			foundMatchingArch = true
			break
		}
	}
	if !foundMatchingArch {
		t.Error("View has the archetype, but the entity count (len) is incorrect")
	}

	beforeCount := len(v.BakedTables)

	var b2 comp.Blueprint
	b2.Comp(posTypeInfo)
	f2 := em.CreateFactory(b2)
	f2.Create(1)
	f2.Next()

	if len(v.BakedTables) != beforeCount {
		t.Errorf("View incorrectly added an archetype that does not satisfy the TagA requirement")
	}
}
