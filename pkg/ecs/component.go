package ecs

import "reflect"

type ComponentID int

type componentsManager struct {
	typeIDs  map[reflect.Type]ComponentID
	storages map[ComponentID]any
	deleters map[ComponentID]func(Entity)
}

func newComponentManager() *componentsManager {
	return &componentsManager{
		typeIDs:  make(map[reflect.Type]ComponentID),
		storages: make(map[ComponentID]any),
		deleters: make(map[ComponentID]func(Entity)),
	}
}

func (m *componentsManager) register(t reflect.Type, createStorage func() any, deleter func(Entity)) ComponentID {
	if id, ok := m.typeIDs[t]; ok {
		return id
	}

	id := ComponentID(len(m.typeIDs))
	m.typeIDs[t] = id
	m.storages[id] = createStorage()
	m.deleters[id] = deleter

	return id
}

func (m *componentsManager) id(t reflect.Type) (ComponentID, bool) {
	id, ok := m.typeIDs[t]
	return id, ok
}

func (m *componentsManager) removeAll(e Entity, mask Bitmask) {
	mask.ForEachSet(func(id ComponentID) {
		if deleter, ok := m.deleters[id]; ok {
			deleter(e)
		}
	})
}

func componentId[T any](m *componentsManager) (ComponentID, bool) {
	componentType := reflect.TypeFor[T]()
	return m.id(componentType)
}

func setComponentValue[T any](m *componentsManager, e Entity, id ComponentID, val T) {
	storage := m.storages[id].(map[Entity]*T)
	v := val
	storage[e] = &v
}

func ensureComponentRegistered[T any](reg *componentsManager) ComponentID {
	componentType := reflect.TypeFor[T]()

	if id, ok := reg.id(componentType); ok {
		return id
	}

	return reg.register(componentType,
		func() any { return make(map[Entity]*T) },
		func(e Entity) {
			if storage, ok := reg.storages[reg.typeIDs[componentType]].(map[Entity]*T); ok {
				delete(storage, e)
			}
		},
	)
}
