package comp

import (
	"fmt"
	"reflect"
)

// DefIndex maps Go types to stable [Def] descriptors.
// Component registration is sequential and deterministic — the first registered
// type gets ID 0, the next gets ID 1, and so on.
type DefIndex struct {
	typeIndex map[reflect.Type]Def
	idIndex   [MaxComponents]Def
}

func (r *DefIndex) Init() {
	r.typeIndex = make(map[reflect.Type]Def)
}

func (r *DefIndex) Reset() {
	if r.typeIndex == nil {
		r.typeIndex = make(map[reflect.Type]Def)
	} else {
		clear(r.typeIndex)
	}
	r.idIndex = [MaxComponents]Def{}
}

// Intern interns a Go type as a component and returns its Def.
// Calling Intern twice for the same type returns the same Def.
func (r *DefIndex) Intern(t reflect.Type) Def {
	if info, ok := r.typeIndex[t]; ok {
		return info
	}

	if len(r.typeIndex) >= MaxComponents {
		panic(fmt.Sprintf("too many components registered: MaxComponents (%d) limit reached", MaxComponents))
	}

	id := ID(len(r.typeIndex))
	info := Def{
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
func (r *DefIndex) ByType(t reflect.Type) (Def, bool) {
	if info, ok := r.typeIndex[t]; ok {
		return info, true
	}
	return Def{}, false
}

// ByID looks up a registered component by its ID.
func (r *DefIndex) ByID(id ID) Def {
	return r.idIndex[id]
}
