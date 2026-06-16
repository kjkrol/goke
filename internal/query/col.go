package query

import (
	"reflect"

	"github.com/kjkrol/goke/internal/comp"
)

// Col is a typed column handle for a tracked component.
// Declare one per component in each system struct, then pass col.Track() to NewView.
// Use col.Slice(v) in All-mode and col.At(v) in Filter-mode instead of the indexed forms.
//
// A single Col must only be used with one View (or multiple views where the
// component always occupies the same Track position).
type Col[T any] struct {
	idx int
}

// Track returns a BlueprintOpt that registers T as a tracked data column and
// records its index in c for later use by Slice and At.
func (c *Col[T]) Track() comp.BlueprintOpt {
	return func(b *comp.Blueprint, cat *comp.Catalog) error {
		c.idx = len(b.CompInfos)
		meta := cat.Intern(reflect.TypeFor[T]())
		return b.Comp(meta)
	}
}

// Slice returns the component slice for the current All-mode chunk.
func (c *Col[T]) Slice(v *View) []T {
	return slice[T](v, c.idx)
}

// At returns a pointer to the component for the current Filter-mode entity.
func (c *Col[T]) At(v *View) *T {
	return at[T](v, c.idx)
}
