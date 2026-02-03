package goke

import (
	"fmt"
	"reflect"

	"github.com/kjkrol/goke/internal/core"
)

// EntityCreate spawns a new entity within the registry and returns its identifier.
// The entity will have no components assigned initially.
func EntityCreate(e *Engine) Entity {
	return e.registry.CreateEntity()
}

// EntityRemove destroys an entity and recycles its ID. All associated
// components are removed and memory is reclaimed. Returns true if the entity existed.
func EntityRemove(e *Engine, entity Entity) bool {
	return e.registry.RemoveEntity(entity)
}

// EntityAllocateComponent reserves memory for a component of type T and returns a direct pointer.
// It is the most efficient way to add data, enabling in-place modification and
// bypassing temporary copies or potential heap escapes.
func EntityAllocateComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compInfo := core.EnsureComponentRegistered[T](eng.registry.ComponentsRegistry)
	ptr, err := eng.registry.AllocateByID(entity, compInfo)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// EntityAllocateComponentByInfo assigns a component to an entity using pre-cached ComponentInfo.
// This is a high-performance alternative to AllocateComponent, as it skips
// registry lookups and type reflection. It returns a typed pointer for
// direct in-place initialization.
func EntityAllocateComponentByInfo[T any](eng *Engine, entity Entity, compInfo ComponentInfo) (*T, error) {
	ptr, err := eng.registry.AllocateByID(entity, compInfo)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// EntityUpsertComponent reserves memory for a component of type T and returns a direct pointer.
// It is the most efficient way to add data, enabling in-place modification and
// bypassing temporary copies or potential heap escapes.
func EntityUpsertComponent[T any](eng *Engine, entity Entity, value T) error {
	ptr, err := EntityAllocateComponent[T](eng, entity)
	if err != nil {
		return err
	}
	*ptr = value
	return nil
}

// EntityGetComponent returns a typed pointer to an entity's component of type T.
func EntityGetComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compType := reflect.TypeFor[T]()
	compInfo, ok := eng.registry.ComponentsRegistry.Get(compType)
	if !ok {
		return nil, fmt.Errorf("component doesn't exist")
	}
	data, err := eng.registry.ComponentGet(entity, compInfo.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}

// EntityGetComponentByType returns a typed pointer to an entity's component of type T.
func EntityGetComponentByType[T any](eng *Engine, entity Entity, compInfo ComponentInfo) (*T, error) {
	data, err := eng.registry.ComponentGet(entity, compInfo.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}

// EntityRemoveComponentByID removes a component from an entity using its ComponentInfo.
func (e *Engine) EntityRemoveComponentByID(entity Entity, compInfo ComponentInfo) error {
	return e.registry.UnassignByID(entity, compInfo)
}

// EntityRemoveComponent removes a component of type T from an entity.
func EntityRemoveComponent[T any](eng *Engine, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	compInfo, ok := eng.registry.ComponentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("component doesn't exist")
	}

	return eng.registry.UnassignByID(entity, compInfo)
}
