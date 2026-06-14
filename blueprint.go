package goke

import (
	"reflect"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/reg"
)

// BlueprintOption configures a View's entity filter by including or excluding component types.
type BlueprintOption func(*comp.Blueprint, *reg.Registry) error

// Include adds a required component type T to the View's filter.
// Only entities that possess this component will be matched.
func Include[T any]() BlueprintOption {
	return func(b *comp.Blueprint, r *reg.Registry) error {
		compMeta := r.CompCatalog.Register(reflect.TypeFor[T]())
		return b.Tag(compMeta.ID)
	}
}

// Exclude adds an exclusion for component type T to the View's filter.
// Entities that possess this component will not be matched.
func Exclude[T any]() BlueprintOption {
	return func(b *comp.Blueprint, r *reg.Registry) error {
		compMeta := r.CompCatalog.Register(reflect.TypeFor[T]())
		return b.Exclude(compMeta.ID)
	}
}
