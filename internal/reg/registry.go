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
	CompCatalog   comp.Catalog
	ViewCatalog   query.Catalog
}

func (r *Registry) Init(cfg Config) {
	validateConst()
	r.CompCatalog.Init()
	r.ViewCatalog.Init(&r.CompCatalog, &r.EntityManager, cfg.View)
	r.EntityManager.Init(cfg.Entity, r.ViewCatalog.OnArchetypeCreated)
}

func (r *Registry) RegCompType(componentType reflect.Type) comp.Meta {
	return r.CompCatalog.Intern(componentType)
}

func (r *Registry) CreateEntity() uid.UID64 {
	return r.EntityManager.Create()
}

func (r *Registry) RemoveEntity(entityID uid.UID64) bool {
	return r.EntityManager.Remove(entityID)
}

func (r *Registry) GetComp(entityID uid.UID64, compID comp.ID) (unsafe.Pointer, error) {
	return r.EntityManager.GetComp(entityID, compID)
}

func (r *Registry) UpsertComp(entityID uid.UID64, compMeta comp.Meta) (unsafe.Pointer, error) {
	return r.EntityManager.UpsertComp(entityID, compMeta)
}

func (r *Registry) RemoveComp(entityID uid.UID64, compMeta comp.Meta) error {
	return r.EntityManager.RemoveComp(entityID, compMeta)
}

func (r *Registry) AddView(blueprint *comp.Blueprint) *query.View {
	return r.ViewCatalog.AddView(blueprint)
}

func (r *Registry) NewView(opts ...comp.BlueprintOpt) *query.View {
	var s comp.Spec0
	s.Init(&r.CompCatalog, opts...)
	return r.AddView(&s.Blueprint)
}

func (r *Registry) Reset() {
	r.EntityManager.Reset()
	r.CompCatalog.Reset()
	r.ViewCatalog.Reset()
}

func validateConst() {
	if arch.HashSize == 0 || (arch.HashSize&(arch.HashSize-1)) != 0 {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}
