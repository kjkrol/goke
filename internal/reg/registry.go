package reg

import (
	"io"
	"os"
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/internal/persist"
	"github.com/kjkrol/goke/v3/internal/query"
)

type Registry struct {
	EntityManager  ent.Manager
	CompDefIndex   comp.DefIndex
	MatcherCatalog query.Catalog
	sharedRemover  *ent.Remover
	paused         bool
	saving         bool
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
		def := spec.AddDefs[i]
		if def.Size == 0 {
			if err := accessSpec.Tag(def.ID); err != nil {
				panic(err)
			}
			continue
		}
		if err := accessSpec.Comp(def); err != nil {
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

// CreateRemover returns a shared Remover instance, built lazily on first
// call and reused thereafter — Remover carries no per-call configuration,
// so one instance serves every caller.
func (r *Registry) CreateRemover() *ent.Remover {
	if r.sharedRemover == nil {
		r.sharedRemover = ent.NewRemover(&r.EntityManager.AddressBook, &r.EntityManager.ArchCatalog)
	}
	return r.sharedRemover
}

// Remover satisfies orch.Mutator — same shared instance as CreateRemover,
// returned as a bulk.Migrator for Scheduler.Register to auto-wire into
// CmdBuf.SetRemover.
func (r *Registry) Remover() bulk.Migrator {
	return r.CreateRemover()
}

func (r *Registry) CreateValueEditor(opts ...comp.EditOpt) *ent.ValueEditor {
	var spec comp.EditSpec
	spec.Init(&r.CompDefIndex, opts...)
	if len(spec.AddDefs) != 1 {
		panic("goke: ValueEditor requires exactly one added component")
	}
	return ent.NewValueEditor(&r.EntityManager.AddressBook, &r.EntityManager.ArchCatalog, spec)
}

// Pause stops Tick from running — a subsequent call panics until Resume.
// General-purpose (a host can use it as an ordinary game pause), and also
// the required precondition for Save: nothing may mutate the world while a
// snapshot is being written. Idempotent — calling it while already paused
// is a no-op.
func (r *Registry) Pause() { r.paused = true }

// Resume clears the paused state set by Pause, allowing Tick again.
// Idempotent — calling it while not paused is a no-op.
func (r *Registry) Resume() { r.paused = false }

// Paused reports whether the registry is currently paused.
func (r *Registry) Paused() bool { return r.paused }

// saveTo writes a full snapshot of the world to w — see [persist.Save].
func (r *Registry) saveTo(w io.Writer) error {
	r.saving = true
	defer func() { r.saving = false }()

	return persist.Save(w, &r.CompDefIndex, &r.EntityManager.AddressBook, &r.EntityManager.ArchCatalog)
}

// Save writes a full snapshot of the world to the file at path. Requires a
// prior Pause (panics otherwise), since nothing may mutate the world while
// the snapshot is being written.
func (r *Registry) Save(path string) error {
	if !r.paused {
		panic("goke: Save called without a prior Pause()")
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return r.saveTo(f)
}

// loadFrom reads a snapshot written by Save, registering components via
// comps — see [persist.Load].
func (r *Registry) loadFrom(rd io.Reader, comps []CompToken) error {
	requests := make([]persist.CompRequest, len(comps))
	for i, c := range comps {
		c := c
		requests[i] = persist.CompRequest{
			Name: c.Name,
			Register: func(wantSize *uint32) error {
				return c.Register(r, wantSize)
			},
		}
	}

	return persist.Load(rd, &r.CompDefIndex, &r.EntityManager.AddressBook, &r.EntityManager.ArchCatalog, requests)
}

// Load reads a snapshot written by Save from the file at path, rebuilding
// its components, archetypes, and entities — with their original IDs. Must
// be called before any component registration (panics otherwise) — Load
// registers components itself, in the file's recorded order, matching the
// given comps by name (see [LoadComp], [CompProvider]).
func (r *Registry) Load(path string, comps []CompToken) error {
	if r.CompDefIndex.Count() > 0 {
		panic("goke: Load called after other components were already registered — Load must run first, before Setup and before any RegComp call")
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return r.loadFrom(f, comps)
}

// Reset clears all entities, components, and system state, returning the
// registry to its initial (post-Init) condition. Also clears the paused
// state. Panics if called while a Save is in progress.
func (r *Registry) Reset() {
	if r.saving {
		panic("goke: Reset called while a Save is in progress")
	}
	r.EntityManager.Reset()
	r.CompDefIndex.Reset()
	r.MatcherCatalog.Reset()
	r.paused = false
}

func validateConst(hashSize uint64) {
	if !isPowerOfTwo(hashSize) {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}

func isPowerOfTwo(n uint64) bool {
	return n > 0 && n&(n-1) == 0
}
