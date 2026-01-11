package ecs

import "reflect"

type ComponentID int

type componentsRegistry struct {
	typeIDs  map[reflect.Type]ComponentID
	storages map[ComponentID]storageInterface
	deleters map[ComponentID]func(Entity)
}

func newComponentsRegistry() *componentsRegistry {
	return &componentsRegistry{
		typeIDs:  make(map[reflect.Type]ComponentID),
		storages: make(map[ComponentID]storageInterface),
		deleters: make(map[ComponentID]func(Entity)),
	}
}

func (m *componentsRegistry) register(t reflect.Type, storage storageInterface, deleter func(Entity)) ComponentID {
	if id, ok := m.typeIDs[t]; ok {
		return id
	}

	id := ComponentID(len(m.typeIDs))
	m.typeIDs[t] = id
	m.storages[id] = storage
	m.deleters[id] = deleter

	return id
}

func (m *componentsRegistry) id(t reflect.Type) (ComponentID, bool) {
	id, ok := m.typeIDs[t]
	return id, ok
}

func (m *componentsRegistry) removeAll(e Entity, mask Bitmask) {
	mask.ForEachSet(func(id ComponentID) {
		if deleter, ok := m.deleters[id]; ok {
			deleter(e)
		}
	})
}

func componentId[T any](m *componentsRegistry) (ComponentID, bool) {
	componentType := reflect.TypeFor[T]()
	return m.id(componentType)
}

func setComponentValue[T any](m *componentsRegistry, entity Entity, id ComponentID, val T) {
	storage := GetTypedStorage[T](m, id)
	storage.Set(entity, val)
}

func ensureComponentRegistered[T any](m *componentsRegistry) ComponentID {
	componentType := reflect.TypeFor[T]()

	if id, ok := m.id(componentType); ok {
		return id
	}

	storage := &ComponentStorage[T]{
		data: make(map[Entity]*T),
	}
	return m.register(componentType, storage, storage.remove)
}

func GetTypedStorage[T any](m *componentsRegistry, id ComponentID) *ComponentStorage[T] {
	s := m.storages[id]
	return s.(*ComponentStorage[T])
}
