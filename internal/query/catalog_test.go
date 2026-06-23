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

func TestMatcherReactivity(t *testing.T) {
	cat, cc, em := newQueryCatalog()

	type Position struct{ X, Y float32 }
	type TagA struct{}

	posTypeInfo := cc.Intern(reflect.TypeFor[Position]())
	tagTypeInfo := cc.Intern(reflect.TypeFor[TagA]())

	var accessSpec comp.AccessSpec
	accessSpec.Comp(posTypeInfo)
	accessSpec.Tag(tagTypeInfo.ID)
	m := cat.AddMatcher(&accessSpec)

	if len(m.BakedTables) != 0 {
		t.Errorf("Expected 0 baked tables initially, got %d", len(m.BakedTables))
	}

	var accessSpec1 comp.AccessSpec
	accessSpec1.Comp(posTypeInfo)
	accessSpec1.Tag(tagTypeInfo.ID)
	f1 := em.CreateFactory(accessSpec1)
	f1.Create(1)
	f1.Next()

	if len(m.BakedTables) == 0 {
		t.Fatal("Matcher failed to detect new archetype created after query initialization!")
	}

	foundMatchingArch := false
	for _, accessSpec := range m.BakedTables {
		if accessSpec.Table.Len() == 1 {
			foundMatchingArch = true
			break
		}
	}
	if !foundMatchingArch {
		t.Error("Matcher has the archetype, but the entity count (len) is incorrect")
	}

	beforeCount := len(m.BakedTables)

	var accessSpec2 comp.AccessSpec
	accessSpec2.Comp(posTypeInfo)
	f2 := em.CreateFactory(accessSpec2)
	f2.Create(1)
	f2.Next()

	if len(m.BakedTables) != beforeCount {
		t.Errorf("Matcher incorrectly added an archetype that does not satisfy the TagA requirement")
	}
}
