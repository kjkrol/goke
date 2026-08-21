package comp

import (
	"fmt"
	"log"
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
//
// t must be encodable (see [ValidateEncodable]) — a bool, a numeric kind, a
// string, a struct, a fixed-size array, or a type implementing
// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler, recursively.
// Ptr, UnsafePointer, Slice, Map, Interface, Chan, and Func fields panic
// unless covered by that escape hatch. A field resolved via string or
// BinaryMarshaler requires a dereference outside the archetype's contiguous
// chunk memory during iteration — Intern logs this once per type.
func (r *DefIndex) Intern(t reflect.Type) Def {
	if info, ok := r.typeIndex[t]; ok {
		return info
	}

	if err := ValidateEncodable(t); err != nil {
		panic(fmt.Sprintf("comp: cannot register component %s: %v", t, err))
	}

	if len(r.typeIndex) >= MaxComponents {
		panic(fmt.Sprintf("too many components registered: MaxComponents (%d) limit reached", MaxComponents))
	}

	for _, path := range OffChunkFields(t) {
		log.Printf("comp: component %s: %s requires a dereference outside the archetype's contiguous chunk memory during iteration, degrading cache locality — consider a fixed-size alternative if this component is iterated in a hot loop", t, path)
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

// Count returns the number of registered component types.
func (r *DefIndex) Count() int {
	return len(r.typeIndex)
}
