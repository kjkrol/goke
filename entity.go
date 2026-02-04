package goke

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

// EntityEnsureComponent returns a direct pointer to the entity's component data.
// If the component already exists, it returns a pointer to the existing data (Update).
// If it doesn't exist, it allocates memory and returns a pointer to the new instance (Insert).
// This is the most efficient way to perform upserts as it enables direct in-place
// modification and avoids temporary copies.
func EntityEnsureComponent[T any](eng *Engine, entity Entity, compType ComponentType) (*T, error) {
	ptr, err := eng.registry.AllocateByID(entity, compType)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// EntityGetComponent returns a typed pointer to an entity's component of type T.
func EntityGetComponent[T any](eng *Engine, entity Entity, compType ComponentType) (*T, error) {
	data, err := eng.registry.ComponentGet(entity, compType.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}

// EntityRemoveComponent removes a component from an entity using its ComponentInfo.
func EntityRemoveComponent(e *Engine, entity Entity, compType ComponentType) error {
	return e.registry.UnassignByID(entity, compType)
}
