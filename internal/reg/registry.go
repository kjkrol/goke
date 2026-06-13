package reg

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/query"
)

type Registry struct {
	EntityPool   uid.UID64Pool
	CompCatalog  comp.Catalog
	ViewRegistry query.Registry
	ArchCatalog  arch.Catalog
}

func (r *Registry) Init(cfg Config) {
	validateConst()
	r.EntityPool = uid.NewUID64Pool(cfg.InitialEntityCap, cfg.FreeIndicesCap)
	r.CompCatalog = comp.NewCatalog()
	r.ViewRegistry = query.NewRegistry(cfg.ViewRegistryInitCap)
	r.ArchCatalog.Init(&r.ViewRegistry, cfg.InitialEntityCap)
}

func (r *Registry) CreateEntity() uid.UID64 {
	entityID := r.EntityPool.Next()
	r.ArchCatalog.AddEntity(entityID, arch.RootID)
	return entityID
}

func (r *Registry) RemoveEntity(entityID uid.UID64) bool {
	if !r.EntityPool.IsValid(entityID) {
		return false
	}

	r.ArchCatalog.UnlinkEntity(entityID)
	r.EntityPool.Release(entityID)
	return true
}

func (r *Registry) UpsertComp(entityID uid.UID64, compMeta comp.Meta) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entityID) {
		return nil, errInvalidEntity
	}

	return r.ArchCatalog.UpsertComp(entityID, compMeta)
}

func (r *Registry) RemoveComp(entityID uid.UID64, compMeta comp.Meta) error {
	if !r.EntityPool.IsValid(entityID) {
		return errInvalidEntity
	}

	r.ArchCatalog.RemoveComp(entityID, compMeta)
	return nil
}

func (r *Registry) RegCompType(componentType reflect.Type) comp.Meta {
	return r.CompCatalog.Register(componentType)
}

func (r *Registry) GetComp(entityID uid.UID64, compID comp.ID) (unsafe.Pointer, error) {
	link, ok := r.ArchCatalog.EntityIndex.Get(entityID)
	if !ok {
		return nil, errInvalidEntity
	}

	a := &r.ArchCatalog.Archetypes[link.ArchId]
	col := a.Table.GetColumn(compID)
	if col == nil {
		return nil, errComponentMissing
	}
	chunk := a.Table.GetChunk(link.Pos.ChunkIdx)
	return col.At(chunk, link.Pos.ChunkSlot), nil
}

func (r *Registry) Reset() {
	r.ArchCatalog.Reset()
	r.CompCatalog.Reset()
	r.ViewRegistry.Reset()
	r.EntityPool.Reset()
}

func validateConst() {
	if arch.HashSize == 0 || (arch.HashSize&(arch.HashSize-1)) != 0 {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}
