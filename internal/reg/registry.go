package reg

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
	"github.com/kjkrol/goke/v2/internal/query"
)

type Registry struct {
	EntityManager  ent.Manager
	CompDefIndex   comp.DefIndex
	MatcherCatalog query.Catalog
}

func (r *Registry) Init(cfg Config) {
	validateConst()
	r.CompDefIndex.Init()
	r.MatcherCatalog.Init(&r.CompDefIndex, &r.EntityManager.AddressBook.Index, &r.EntityManager.ArchCatalog, cfg.Matcher)
	r.EntityManager.Init(cfg.Entity, r.MatcherCatalog.OnArchetypeCreated)
}

func (r *Registry) RegComp(compType reflect.Type) comp.ID {
	return r.CompDefIndex.Intern(compType).ID
}

func (r *Registry) CreateFactory(opts ...comp.EditOpt) *ent.Factory {
	var spec comp.EditSpec
	spec.Init(&r.CompDefIndex, opts...)
	if len(spec.DelDefs) > 0 {
		panic("goke: Factory cannot remove components — use Add only")
	}
	var accessSpec comp.AccessSpec
	for i := range spec.AddDefs {
		if err := accessSpec.Comp(spec.AddDefs[i]); err != nil {
			panic(err)
		}
	}
	return r.EntityManager.CreateFactory(accessSpec)
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

func (r *Registry) AddMatcher(opts ...comp.AccessOpt) *query.Matcher {
	var accessSpec comp.AccessSpec
	accessSpec.Init(&r.CompDefIndex, opts...)
	return r.MatcherCatalog.AddMatcher(&accessSpec)
}

func (r *Registry) CreateEditor(opts ...comp.EditOpt) *ent.Editor {
	var spec comp.EditSpec
	spec.Init(&r.CompDefIndex, opts...)
	return r.EntityManager.CreateEditor(spec)
}

func (r *Registry) Reset() {
	r.EntityManager.Reset()
	r.CompDefIndex.Reset()
	r.MatcherCatalog.Reset()
}

func validateConst() {
	if arch.HashSize == 0 || (arch.HashSize&(arch.HashSize-1)) != 0 {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}
