package core

import (
	"fmt"
	"reflect"
)

type ComponentID uint8

type ComponentInfo struct {
	ID   ComponentID
	Size uintptr
	Type reflect.Type
}

type ComponentsRegistry struct {
	typeToInfo map[reflect.Type]ComponentInfo
	idToInfo   [MaxComponents]ComponentInfo
}

func (r *ComponentsRegistry) Reset() {
	if r.typeToInfo == nil {
		r.typeToInfo = make(map[reflect.Type]ComponentInfo)
	} else {
		clear(r.typeToInfo)
	}
	r.idToInfo = [MaxComponents]ComponentInfo{}
}

func NewComponentsRegistry() ComponentsRegistry {
	return ComponentsRegistry{
		typeToInfo: make(map[reflect.Type]ComponentInfo),
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

	idInt := len(r.typeToInfo)

	if idInt >= MaxComponents {
		panic(fmt.Sprintf("too many components registered: MaxComponents (%d) limit reached", MaxComponents))
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
