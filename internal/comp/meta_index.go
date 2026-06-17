package comp

import (
	"fmt"
	"reflect"
)

// MetaIndex maps Go types to stable [Meta] descriptors.
// Component registration is sequential and deterministic — the first registered
// type gets ID 0, the next gets ID 1, and so on.
type MetaIndex struct {
	typeIndex map[reflect.Type]Meta
	idIndex   [MaxComponents]Meta
}

func (r *MetaIndex) Init() {
	r.typeIndex = make(map[reflect.Type]Meta)
}

func (r *MetaIndex) Reset() {
	if r.typeIndex == nil {
		r.typeIndex = make(map[reflect.Type]Meta)
	} else {
		clear(r.typeIndex)
	}
	r.idIndex = [MaxComponents]Meta{}
}

// Intern interns a Go type as a component and returns its Meta.
// Calling Intern twice for the same type returns the same Meta.
func (r *MetaIndex) Intern(t reflect.Type) Meta {
	if info, ok := r.typeIndex[t]; ok {
		return info
	}

	if len(r.typeIndex) >= MaxComponents {
		panic(fmt.Sprintf("too many components registered: MaxComponents (%d) limit reached", MaxComponents))
	}

	id := ID(len(r.typeIndex))
	info := Meta{
		ID:    id,
		Size:  t.Size(),
		Align: uintptr(t.Align()),
		Type:  t,
	}

	r.typeIndex[t] = info
	r.idIndex[id] = info
	return info
}

// ByType looks up a registered component by its Go type.
func (r *MetaIndex) ByType(t reflect.Type) (Meta, bool) {
	if info, ok := r.typeIndex[t]; ok {
		return info, true
	}
	return Meta{}, false
}

// ByID looks up a registered component by its ID.
func (r *MetaIndex) ByID(id ID) Meta {
	return r.idIndex[id]
}
