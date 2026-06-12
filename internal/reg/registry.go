package reg

import (
	"errors"
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/internal/mem"
	"github.com/kjkrol/goke/internal/view"
)

var (
	errInvalidEntity    = errors.New("invalid entity")
	errComponentMissing = errors.New("component not found in archetype")
)

type Registry struct {
	EntityPool         uid.UID64Pool
	ComponentsRegistry core.ComponentsRegistry
	ViewRegistry       view.ViewRegistry
	ArchetypeRegistry  *arch.ArchetypeRegistry
}

var _ core.ComponentReader = (*Registry)(nil)

func NewRegistry(cfg RegistryConfig) *Registry {
	validateConst()
	r := &Registry{
		EntityPool:         uid.NewUID64Pool(cfg.InitialEntityCap, cfg.FreeIndicesCap),
		ComponentsRegistry: core.NewComponentsRegistry(),
		ViewRegistry:       view.NewViewRegistry(cfg.ViewRegistryInitCap),
	}
	r.ArchetypeRegistry = arch.NewArchetypeRegistry(&r.ComponentsRegistry, &r.ViewRegistry, cfg.InitialEntityCap)
	return r
}

func NewView(blueprint *Blueprint, layout []core.ComponentInfo, r *Registry) *view.View {
	var includeMask, excludeMask core.ArchetypeMask

	for _, info := range blueprint.compInfos {
		includeMask = includeMask.Set(info.ID)
	}
	for _, id := range blueprint.tagIDs {
		includeMask = includeMask.Set(id)
	}
	for _, id := range blueprint.exCompIDs {
		if includeMask.IsSet(id) {
			panic("ECS View Error: component cannot be both REQUIRED and EXCLUDED")
		}
		excludeMask = excludeMask.Set(id)
	}

	return view.NewView(includeMask, excludeMask, layout, r.ArchetypeRegistry, &r.ViewRegistry)
}

func (r *Registry) CreateEntity() uid.UID64 {
	entity := r.EntityPool.Next()
	r.ArchetypeRegistry.AddEntity(entity, core.RootArchetypeId)
	return entity
}

func (r *Registry) RemoveEntity(entity uid.UID64) bool {
	if !r.EntityPool.IsValid(entity) {
		return false
	}

	r.ArchetypeRegistry.UnlinkEntity(entity)
	r.EntityPool.Release(entity)
	return true
}

func (r *Registry) AllocateByID(entity uid.UID64, compInfo core.ComponentInfo) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entity) {
		return nil, errInvalidEntity
	}

	return r.ArchetypeRegistry.AllocateComponentMemory(entity, compInfo)
}

func (r *Registry) UnassignByID(entity uid.UID64, compInfo core.ComponentInfo) error {
	if !r.EntityPool.IsValid(entity) {
		return errInvalidEntity
	}

	r.ArchetypeRegistry.UnAssign(entity, compInfo)
	return nil
}

func (r *Registry) RegisterComponentType(componentType reflect.Type) core.ComponentInfo {
	return r.ComponentsRegistry.GetOrRegister(componentType)
}

func (r *Registry) ComponentGet(entity uid.UID64, compID core.ComponentID) (unsafe.Pointer, error) {
	link, ok := r.ArchetypeRegistry.EntityLinkStore.Get(entity)
	if !ok {
		return nil, errInvalidEntity
	}

	a := &r.ArchetypeRegistry.Archetypes[link.ArchId]
	localIdx := a.Map[compID]
	if localIdx == mem.InvalidLocalID {
		return nil, errComponentMissing
	}
	col := &a.Columns[localIdx]
	return unsafe.Add(a.Memory.Pages[link.PageIdx].Ptr, col.PageOffset+uintptr(link.PageSlot)*col.ItemSize), nil
}

func (r *Registry) Reset() {
	r.ArchetypeRegistry.Reset()
	r.ComponentsRegistry.Reset()
	r.ViewRegistry.Reset()
	r.EntityPool.Reset()
}

func validateConst() {
	if core.HashSize == 0 || (core.HashSize&(core.HashSize-1)) != 0 {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}
