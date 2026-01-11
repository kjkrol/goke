package ecs

type Registry struct {
	entitiesRegistry   *entitiesRegistry
	componentsRegistry *componentsRegistry
}

func newRegistry() *Registry {
	return &Registry{
		entitiesRegistry:   newEntitiesRegistry(),
		componentsRegistry: newComponentsRegistry(),
	}
}

func (r *Registry) RemoveEntity(entity Entity) {
	mask, ok := r.entitiesRegistry.mask(entity)
	if !ok {
		return
	}

	r.componentsRegistry.removeAll(entity, mask)
	r.entitiesRegistry.destroy(entity)
}

func assign[T any](reg *Registry, entity Entity, component T) {
	id := ensureComponentRegistered[T](reg.componentsRegistry)
	assignByID(reg, entity, id, component)
}

func assignByID[T any](reg *Registry, entity Entity, id ComponentID, component T) {
	mask, ok := reg.entitiesRegistry.mask(entity)
	if !ok {
		return
	}
	setComponentValue(reg.componentsRegistry, entity, id, component)
	reg.entitiesRegistry.updateMask(entity, mask.Set(id))
}

func unassign[T any](reg *Registry, entity Entity) {
	id, ok := componentId[T](reg.componentsRegistry)
	if !ok {
		return
	}

	reg.UnassignByID(entity, id)
}

func (r *Registry) UnassignByID(e Entity, id ComponentID) {
	if deleter, ok := r.componentsRegistry.deleters[id]; ok {
		deleter(e)
	}

	if mask, ok := r.entitiesRegistry.mask(e); ok {
		newMask := mask.Clear(id)
		r.entitiesRegistry.updateMask(e, newMask)
	}
}

func getComponent[T any](reg *Registry, e Entity) *T {
	id, ok := componentId[T](reg.componentsRegistry)
	if !ok {
		return nil
	}

	mask, exists := reg.entitiesRegistry.mask(e)
	if !exists || !mask.IsSet(id) {
		return nil
	}

	storage := reg.componentsRegistry.storages[id].(*ComponentStorage[T])
	return storage.data[e]
}
