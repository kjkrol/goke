package comp

import (
	"reflect"

	"github.com/kjkrol/goke/v2/iter"
)

// AccessOpt configures a AccessSpec's entity filter by including or excluding component types.
type AccessOpt func(*AccessSpec, *DefIndex) error

// Include adds a required component type T to the AccessSpec's filter.
func Include[T any]() AccessOpt {
	return func(s *AccessSpec, mi *DefIndex) error {
		compDef := mi.Intern(reflect.TypeFor[T]())
		return s.Tag(compDef.ID)
	}
}

// Track returns a AccessOpt that registers T as a tracked data column and
// sets col.Idx to its position. Pass col.Slice or col.At to access data.
// The same opt may be reused across multiple views as long as T occupies the
// same track position in each.
func Track[T any](col *iter.ArrayRef[T]) AccessOpt {
	return func(s *AccessSpec, mi *DefIndex) error {
		col.Idx = len(s.CompInfos)
		compDef := mi.Intern(reflect.TypeFor[T]())
		return s.Comp(compDef)
	}
}

// Exclude adds an exclusion for component type T to the AccessSpec's filter.
func Exclude[T any]() AccessOpt {
	return func(s *AccessSpec, mi *DefIndex) error {
		compDef := mi.Intern(reflect.TypeFor[T]())
		return s.Exclude(compDef.ID)
	}
}
