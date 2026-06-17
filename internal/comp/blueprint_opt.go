package comp

import (
	"reflect"

	"github.com/kjkrol/goke/iter"
)

// BlueprintOpt configures a Blueprint's entity filter by including or excluding component types.
type BlueprintOpt func(*Blueprint, *MetaIndex) error

// Include adds a required component type T to the Blueprint's filter.
func Include[T any]() BlueprintOpt {
	return func(b *Blueprint, mi *MetaIndex) error {
		compMeta := mi.Intern(reflect.TypeFor[T]())
		return b.Tag(compMeta.ID)
	}
}

// Track returns a BlueprintOpt that registers T as a tracked data column and
// sets col.Idx to its position. Pass col.Slice or col.At to access data.
// The same opt may be reused across multiple views as long as T occupies the
// same track position in each.
func Track[T any](col *iter.Col[T]) BlueprintOpt {
	return func(b *Blueprint, mi *MetaIndex) error {
		col.Idx = len(b.CompInfos)
		compMeta := mi.Intern(reflect.TypeFor[T]())
		return b.Comp(compMeta)
	}
}

// Exclude adds an exclusion for component type T to the Blueprint's filter.
func Exclude[T any]() BlueprintOpt {
	return func(b *Blueprint, mi *MetaIndex) error {
		compMeta := mi.Intern(reflect.TypeFor[T]())
		return b.Exclude(compMeta.ID)
	}
}
