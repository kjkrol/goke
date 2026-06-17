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
	CompMetaIndex comp.MetaIndex
	ViewCatalog   query.Catalog
}

func (r *Registry) Init(cfg Config) {
	validateConst()
	r.CompMetaIndex.Init()
	r.ViewCatalog.Init(&r.CompMetaIndex, &r.EntityManager.AddressBook.Index, &r.EntityManager.ArchCatalog, cfg.View)
	r.EntityManager.Init(cfg.Entity, r.ViewCatalog.OnArchetypeCreated)
}

func (r *Registry) RegCompType(compType reflect.Type) comp.Meta {
	return r.CompMetaIndex.Intern(compType)
}

func (r *Registry) CreateFactory(opts ...comp.BlueprintOpt) *ent.Factory {
	var b comp.Blueprint
	b.Init(&r.CompMetaIndex, opts...)
	return r.EntityManager.CreateFactory(b)
}

func (r *Registry) Remove(entID uid.UID64) bool {
	return r.EntityManager.Remove(entID)
}

func (r *Registry) GetComp(entID uid.UID64, compID comp.ID) (unsafe.Pointer, error) {
	return r.EntityManager.GetComp(entID, compID)
}

func (r *Registry) UpsertComp(entID uid.UID64, compMeta comp.Meta) (unsafe.Pointer, error) {
	return r.EntityManager.UpsertComp(entID, compMeta)
}

func (r *Registry) RemoveComp(entID uid.UID64, compMeta comp.Meta) error {
	return r.EntityManager.RemoveComp(entID, compMeta)
}

func (r *Registry) AddView(opts ...comp.BlueprintOpt) *query.View {
	var b comp.Blueprint
	b.Init(&r.CompMetaIndex, opts...)
	return r.ViewCatalog.AddView(&b)
}

func (r *Registry) Reset() {
	r.EntityManager.Reset()
	r.CompMetaIndex.Reset()
	r.ViewCatalog.Reset()
}

func validateConst() {
	if arch.HashSize == 0 || (arch.HashSize&(arch.HashSize-1)) != 0 {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}
