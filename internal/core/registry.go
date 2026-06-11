package core

import (
	"errors"
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"
)

var (
	errInvalidEntity    = errors.New("invalid entity")
	errComponentMissing = errors.New("component not found in archetype")
)

type Registry struct {
	EntityPool         uid.UID64Pool
	ComponentsRegistry ComponentsRegistry
	ViewRegistry       ViewRegistry
	ArchetypeRegistry  *ArchetypeRegistry
}

var _ ReadOnlyRegistry = (*Registry)(nil)

func NewRegistry(cfg RegistryConfig) *Registry {
	validateConst()
	reg := &Registry{
		EntityPool:         uid.NewUID64Pool(cfg.InitialEntityCap, cfg.FreeIndicesCap),
		ComponentsRegistry: NewComponentsRegistry(),
		ViewRegistry:       NewViewRegistry(cfg.ViewRegistryInitCap),
	}
	reg.ArchetypeRegistry = NewArchetypeRegistry(&reg.ComponentsRegistry, &reg.ViewRegistry, cfg)
	return reg
}

// CreateEntity allocates a new empty entity in the registry.
//
// Deprecated: This method should not be used directly in the public API.
// To ensure consistent state and proper component initialization,
// entities should be created through a ArchetypeEntryBlueprint instead.
func (r *Registry) CreateEntity() uid.UID64 {
	entity := r.EntityPool.Next()
	r.ArchetypeRegistry.AddEntity(entity, RootArchetypeId)
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

// AllocateByID ensures the entity is valid and performs the allocation in the archetype.
func (r *Registry) AllocateByID(entity uid.UID64, compInfo ComponentInfo) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entity) {
		return nil, errInvalidEntity
	}

	return r.ArchetypeRegistry.AllocateComponentMemory(entity, compInfo)
}

func (r *Registry) UnassignByID(entity uid.UID64, compInfo ComponentInfo) error {
	if !r.EntityPool.IsValid(entity) {
		return errInvalidEntity
	}

	r.ArchetypeRegistry.UnAssign(entity, compInfo)
	return nil
}

func (r *Registry) RegisterComponentType(componentType reflect.Type) ComponentInfo {
	return r.ComponentsRegistry.GetOrRegister(componentType)
}

func (r *Registry) ComponentGet(entity uid.UID64, compID ComponentID) (unsafe.Pointer, error) {
	link, ok := r.ArchetypeRegistry.EntityLinkStore.Get(entity)
	if !ok {
		return nil, errInvalidEntity
	}

	arch := &r.ArchetypeRegistry.Archetypes[link.ArchId]
	localIdx := arch.Map[compID]
	if localIdx == InvalidLocalID {
		return nil, errComponentMissing
	}
	col := &arch.Columns[localIdx]
	return unsafe.Add(arch.Memory.Pages[link.PageIdx].Ptr, col.PageOffset+uintptr(link.PageSlot)*col.ItemSize), nil
}

func (r *Registry) Reset() {
	r.ArchetypeRegistry.Reset()
	r.ComponentsRegistry.Reset()
	r.ViewRegistry.Reset()
	r.EntityPool.Reset()
}

func validateConst() {
	if HashSize == 0 || (HashSize&(HashSize-1)) != 0 {
		panic("CRITICAL: HashSize must be a power of 2!")
	}
}
