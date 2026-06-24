package comp

import (
	"fmt"
	"reflect"

	"github.com/kjkrol/goke/iter"
)

// EditSpec is the set of structural changes an Editor applies to an entity:
// components to add (with their column bound for value writes) and to remove.
type EditSpec struct {
	AddDefs []Def
	DelDefs []Def
}

// EditOpt configures an EditSpec — the structural counterpart of [AccessOpt]
// ([AccessOpt] accesses component values; EditOpt adds or removes components).
type EditOpt func(*EditSpec, *DefIndex) error

// Init applies opts to the EditSpec in place. Panics if any opt returns an error.
func (s *EditSpec) Init(mi *DefIndex, opts ...EditOpt) {
	for _, opt := range opts {
		if err := opt(s, mi); err != nil {
			panic(fmt.Sprintf("comp: edit option: %v", err))
		}
	}
}

// Add registers T as a component to add, binding col so its value can be written
// after the edit. col.Idx is set to T's position among the added columns.
func Add[T any](col *iter.ArrayRef[T]) EditOpt {
	return func(s *EditSpec, mi *DefIndex) error {
		col.Idx = len(s.AddDefs)
		s.AddDefs = append(s.AddDefs, mi.Intern(reflect.TypeFor[T]()))
		return nil
	}
}

// Del registers T as a component to remove.
func Del[T any]() EditOpt {
	return func(s *EditSpec, mi *DefIndex) error {
		s.DelDefs = append(s.DelDefs, mi.Intern(reflect.TypeFor[T]()))
		return nil
	}
}
