package comp

import "reflect"

// BlueprintOpt configures a Blueprint's entity filter by including or excluding component types.
type BlueprintOpt func(*Blueprint, *Catalog) error

// Include adds a required component type T to the Blueprint's filter.
func Include[T any]() BlueprintOpt {
	return func(b *Blueprint, c *Catalog) error {
		compMeta := c.Intern(reflect.TypeFor[T]())
		return b.Tag(compMeta.ID)
	}
}

// Exclude adds an exclusion for component type T to the Blueprint's filter.
func Exclude[T any]() BlueprintOpt {
	return func(b *Blueprint, c *Catalog) error {
		compMeta := c.Intern(reflect.TypeFor[T]())
		return b.Exclude(compMeta.ID)
	}
}
