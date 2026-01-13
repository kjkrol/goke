package ecs

import (
	"reflect"
)

type ComponentID int

type ComponentInfo struct {
	ID   ComponentID
	Size uintptr
	Type reflect.Type
}

type componentsRegistry struct {
	typeToInfo map[reflect.Type]*ComponentInfo
	idToInfo   map[ComponentID]*ComponentInfo
}

func newComponentsRegistry() *componentsRegistry {
	return &componentsRegistry{
		typeToInfo: make(map[reflect.Type]*ComponentInfo),
		idToInfo:   make(map[ComponentID]*ComponentInfo),
	}
}

func (m *componentsRegistry) Get(t reflect.Type) (ComponentID, bool) {
	if info, ok := m.typeToInfo[t]; ok {
		return info.ID, true
	}
	return 0, false
}

func (m *componentsRegistry) GetOrRegister(t reflect.Type) ComponentID {
	if info, ok := m.typeToInfo[t]; ok {
		return info.ID
	}

	id := ComponentID(len(m.typeToInfo))
	info := &ComponentInfo{
		ID:   id,
		Size: t.Size(),
		Type: t,
	}

	m.typeToInfo[t] = info
	m.idToInfo[id] = info
	return id
}

func ensureComponentRegistered[T any](m *componentsRegistry) ComponentID {
	componentType := reflect.TypeFor[T]()
	return m.GetOrRegister(componentType)
}
