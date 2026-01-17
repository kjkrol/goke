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

type ComponentsRegistry struct {
	typeToInfo map[reflect.Type]*ComponentInfo
	idToInfo   map[ComponentID]*ComponentInfo
}

func newComponentsRegistry() *ComponentsRegistry {
	return &ComponentsRegistry{
		typeToInfo: make(map[reflect.Type]*ComponentInfo),
		idToInfo:   make(map[ComponentID]*ComponentInfo),
	}
}

func (r *ComponentsRegistry) Get(t reflect.Type) (ComponentID, bool) {
	if info, ok := r.typeToInfo[t]; ok {
		return info.ID, true
	}
	return 0, false
}

func (r *ComponentsRegistry) GetOrRegister(t reflect.Type) ComponentID {
	if info, ok := r.typeToInfo[t]; ok {
		return info.ID
	}

	id := ComponentID(len(r.typeToInfo))
	info := &ComponentInfo{
		ID:   id,
		Size: t.Size(),
		Type: t,
	}

	r.typeToInfo[t] = info
	r.idToInfo[id] = info
	return id
}

func ensureComponentRegistered[T any](m *ComponentsRegistry) ComponentID {
	componentType := reflect.TypeFor[T]()
	return m.GetOrRegister(componentType)
}
