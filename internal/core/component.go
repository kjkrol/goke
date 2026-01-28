package core

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
	typeToInfo map[reflect.Type]ComponentInfo
	idToInfo   map[ComponentID]ComponentInfo
}

func NewComponentsRegistry() *ComponentsRegistry {
	return &ComponentsRegistry{
		typeToInfo: make(map[reflect.Type]ComponentInfo),
		idToInfo:   make(map[ComponentID]ComponentInfo),
	}
}

func (r *ComponentsRegistry) Get(t reflect.Type) (ComponentInfo, bool) {
	if info, ok := r.typeToInfo[t]; ok {
		return info, true
	}
	return ComponentInfo{}, false
}

func (r *ComponentsRegistry) GetOrRegister(t reflect.Type) ComponentInfo {
	if info, ok := r.typeToInfo[t]; ok {
		return info
	}

	id := ComponentID(len(r.typeToInfo))
	info := ComponentInfo{
		ID:   id,
		Size: t.Size(),
		Type: t,
	}

	r.typeToInfo[t] = info
	r.idToInfo[id] = info
	return info
}

func EnsureComponentRegistered[T any](m *ComponentsRegistry) ComponentInfo {
	componentType := reflect.TypeFor[T]()
	return m.GetOrRegister(componentType)
}
