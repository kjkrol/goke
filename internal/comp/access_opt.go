package comp

import (
	"reflect"

	"github.com/kjkrol/goke/v3/iter"
)

// AccessOpt configures a AccessSpec's entity filter by including or excluding component types.
type AccessOpt func(*AccessSpec, *DefIndex) error

func Include[T any]() AccessOpt {
	return func(s *AccessSpec, mi *DefIndex) error {
		compDef := mi.Intern(reflect.TypeFor[T]())
		return s.Tag(compDef.ID)
	}
}

// Track registers T as a tracked data column and sets col.Idx to its position;
// reusable across views as long as T keeps the same track position.
func Track[T any](col *iter.ArrayRef[T]) AccessOpt {
	return func(s *AccessSpec, mi *DefIndex) error {
		col.Idx = len(s.CompInfos)
		compDef := mi.Intern(reflect.TypeFor[T]())
		return s.Comp(compDef)
	}
}

func Exclude[T any]() AccessOpt {
	return func(s *AccessSpec, mi *DefIndex) error {
		compDef := mi.Intern(reflect.TypeFor[T]())
		return s.Exclude(compDef.ID)
	}
}
