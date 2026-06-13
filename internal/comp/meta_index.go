package comp

import (
	"fmt"
	"reflect"
)

type metaIndex struct {
	typeIndex map[reflect.Type]Meta
	idIndex   [MaxComponents]Meta
}

func (r *metaIndex) Init() {
	r.typeIndex = make(map[reflect.Type]Meta)
}

func (r *metaIndex) reset() {
	if r.typeIndex == nil {
		r.typeIndex = make(map[reflect.Type]Meta)
	} else {
		clear(r.typeIndex)
	}
	r.idIndex = [MaxComponents]Meta{}
}

func (r *metaIndex) byType(t reflect.Type) (Meta, bool) {
	if info, ok := r.typeIndex[t]; ok {
		return info, true
	}
	return Meta{}, false
}

func (r *metaIndex) byID(id ID) Meta {
	return r.idIndex[id]
}

func (r *metaIndex) intern(t reflect.Type) Meta {
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
