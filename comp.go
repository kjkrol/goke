package goke

import (
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/iter"
)

// Comp gives typed read/write access to a component's value, and declares
// that component as tracked (data access) or added (structural change) when
// passed to a builder. Declare one as a variable, then pass its address
// directly to NewFactory/NewQueryBuilder/NewEditorBuilder — it binds itself,
// no separate wrapping call needed. Use comp.Slice(cursor) in All/Factory-mode
// and comp.At(cursor) in Pick/Seek-mode.
type Comp[T any] struct {
	col iter.ArrayRef[T]
}

// Slice returns the component slice for the current All-mode chunk or
// Factory batch. Its length equals len(cursor.IDs), so ranging cursor.IDs in
// the inner loop lets the compiler eliminate bounds checks on slice[i] accesses.
func (c *Comp[T]) Slice(cur *Cursor) []T { return c.col.Slice(cur) }

// At returns a pointer to the component for the current Pick/Seek-mode entity.
func (c *Comp[T]) At(cur *Cursor) *T { return c.col.At(cur) }

// Trackable is satisfied by *Comp[T] for any T — it lets NewQueryBuilder
// accept components (&comp) directly as tracked data columns.
type Trackable interface {
	// asTrack is unexported so *Comp[T] is the only implementer — this is a
	// sealed interface, not an extension point.
	asTrack() Opt
}

// Addable is satisfied by *Comp[T] for any T — it lets NewFactory and
// NewEditorBuilder accept components (&comp) directly as added components.
type Addable interface {
	// asAdd is unexported so *Comp[T] is the only implementer — this is a
	// sealed interface, not an extension point.
	asAdd() EditOpt
}

func (c *Comp[T]) asTrack() Opt   { return comp.Track[T](&c.col) }
func (c *Comp[T]) asAdd() EditOpt { return comp.Add[T](&c.col) }
