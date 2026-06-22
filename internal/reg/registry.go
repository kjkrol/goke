package reg

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
	"github.com/kjkrol/goke/internal/query"
)

type Registry struct {
	EntityManager ent.Manager
	CompDefIndex  comp.DefIndex
	ViewCatalog   query.Catalog
}

func (r *Registry) Init(cfg Config) {
	validateConst()
	r.CompDefIndex.Init()
	r.ViewCatalog.Init(&r.CompDefIndex, &r.EntityManager.AddressBook.Index, &r.EntityManager.ArchCatalog, cfg.View)
	r.EntityManager.Init(cfg.Entity, r.ViewCatalog.OnArchetypeCreated)
}

func (r *Registry) RegComp(compType reflect.Type) comp.ID {
	return r.CompDefIndex.Intern(compType).ID
}

func (r *Registry) CreateFactory(opts ...comp.BlueprintOpt) *ent.Factory {
	var b comp.Blueprint
	b.Init(&r.CompDefIndex, opts...)
	return r.EntityManager.CreateFactory(b)
}

func (r *Registry) Remove(entID uid.UID64) bool {
	return r.EntityManager.Remove(entID)
}

func (r *Registry) UpsertComp(entID uid.UID64, compID comp.ID) (unsafe.Pointer, error) {
	return r.EntityManager.UpsertComp(entID, r.CompDefIndex.ByID(compID))
}

func (r *Registry) RemoveComp(entID uid.UID64, compID comp.ID) error {
	return r.EntityManager.RemoveComp(entID, r.CompDefIndex.ByID(compID))
}

func (r *Registry) AddView(opts ...comp.BlueprintOpt) *query.View {
	var b comp.Blueprint
	b.Init(&r.CompDefIndex, opts...)
	return r.ViewCatalog.AddView(&b)
}

func (r *Registry) CreateLookup(opts ...comp.BlueprintOpt) *query.Lookup {
	var b comp.Blueprint
	b.Init(&r.CompDefIndex, opts...)
	lookup := &query.Lookup{}
	lookup.Init(&r.EntityManager.AddressBook.Index, &r.EntityManager.ArchCatalog, b.CompIDs())
	return lookup
}

func (r *Registry) Reset() {
	r.EntityManager.Reset()
	r.CompDefIndex.Reset()
	r.ViewCatalog.Reset()
}

func validateConst() {
	if arch.HashSize == 0 || (arch.HashSize&(arch.HashSize-1)) != 0 {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}
