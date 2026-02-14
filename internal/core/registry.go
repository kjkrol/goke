package core

import (
	"fmt"
	"reflect"
	"unsafe"
)

type Registry struct {
	EntityPool         EntityGenerationalPool
	ComponentsRegistry ComponentsRegistry
	ViewRegistry       ViewRegistry
	ArchetypeRegistry  *ArchetypeRegistry
}

var _ ReadOnlyRegistry = (*Registry)(nil)

func NewRegistry(cfg RegistryConfig) *Registry {
	validateConst()
	reg := &Registry{
		EntityPool:         NewEntityGenerator(cfg.InitialEntityCap, cfg.FreeIndicesCap),
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
func (r *Registry) CreateEntity() Entity {
	entity := r.EntityPool.Next()
	r.ArchetypeRegistry.AddEntity(entity, RootArchetypeId)
	return entity
}

func (r *Registry) RemoveEntity(entity Entity) bool {
	if !r.EntityPool.IsValid(entity) {
		return false
	}

	r.ArchetypeRegistry.UnlinkEntity(entity)
	r.EntityPool.Release(entity)
	return true
}

// AllocateByID ensures the entity is valid and performs the allocation in the archetype.
func (r *Registry) AllocateByID(entity Entity, compInfo ComponentInfo) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entity) {
		return nil, fmt.Errorf("invalid entity")
	}

	return r.ArchetypeRegistry.AllocateComponentMemory(entity, compInfo)
}

func (r *Registry) UnassignByID(entity Entity, compInfo ComponentInfo) error {
	if !r.EntityPool.IsValid(entity) {
		return fmt.Errorf("Invalid Entity")
	}

	r.ArchetypeRegistry.UnAssign(entity, compInfo)
	return nil
}

func (r *Registry) RegisterComponentType(componentType reflect.Type) ComponentInfo {
	return r.ComponentsRegistry.GetOrRegister(componentType)
}

func (r *Registry) ComponentGet(entity Entity, compID ComponentID) (unsafe.Pointer, error) {
	if !r.EntityPool.IsValid(entity) {
		return nil, fmt.Errorf("Invalid Entity")
	}

	link, ok := r.ArchetypeRegistry.EntityLinkStore.Get(entity)
	if !ok {
		return nil, fmt.Errorf("entity not found in registry")
	}

	arch := &r.ArchetypeRegistry.Archetypes[link.ArchId]
	col := arch.GetColumn(compID)
	if col == nil {
		return nil, fmt.Errorf("component not found in archetype")
	}
	chunk := arch.Memory.GetChunk(link.ChunkIdx)

	return col.GetPointer(chunk, link.ChunkRow), nil
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
