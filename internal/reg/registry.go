package reg

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
	"github.com/kjkrol/goke/v2/internal/query"
)

type Registry struct {
	EntityManager  ent.Manager
	CompDefIndex   comp.DefIndex
	MatcherCatalog query.Catalog
	sharedRemover  *ent.Remover
}

func (r *Registry) Init(cfg Config) {
	validateConst(arch.HashSize)
	r.CompDefIndex.Init()
	r.MatcherCatalog.Init(&r.CompDefIndex, &r.EntityManager.AddressBook.Index, &r.EntityManager.ArchCatalog, cfg.Matcher)
	r.EntityManager.Init(cfg.Entity, func(a *arch.Archetype) {
		r.MatcherCatalog.OnArchetypeCreated(a)
	})
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
	return ent.NewEditor(&r.EntityManager.AddressBook, &r.EntityManager.ArchCatalog, spec)
}

// Remover returns a shared Remover instance, built lazily on first call and
// reused thereafter — Remover carries no per-call configuration, so one
// instance serves every caller.
func (r *Registry) Remover() bulk.Migrator {
	if r.sharedRemover == nil {
		r.sharedRemover = ent.NewRemover(&r.EntityManager.AddressBook, &r.EntityManager.ArchCatalog)
	}
	return r.sharedRemover
}

func (r *Registry) CreateValueEditor(opts ...comp.EditOpt) *ent.ValueEditor {
	var spec comp.EditSpec
	spec.Init(&r.CompDefIndex, opts...)
	if len(spec.AddDefs) != 1 {
		panic("goke: ValueEditor requires exactly one added component")
	}
	return ent.NewValueEditor(&r.EntityManager.AddressBook, &r.EntityManager.ArchCatalog, spec)
}

func (r *Registry) Reset() {
	r.EntityManager.Reset()
	r.CompDefIndex.Reset()
	r.MatcherCatalog.Reset()
}

func validateConst(hashSize uint64) {
	if !isPowerOfTwo(hashSize) {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}

func isPowerOfTwo(n uint64) bool {
	return n > 0 && n&(n-1) == 0
}
