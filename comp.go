package goke

import (
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/reg"
	"github.com/kjkrol/goke/v3/iter"
)

// Comp gives typed read/write access to a component. Declare one, pass its
// address (&comp) directly to NewFactory/NewQueryBuilder/NewEditorBuilder —
// it binds itself. Use Slice(cursor) in All/Factory-mode, At(cursor) in
// Pick/Seek-mode.
//
// T must be encodable — recursively a bool, numeric kind, string, struct,
// fixed-size array, or a type implementing encoding.BinaryMarshaler and
// encoding.BinaryUnmarshaler — or RegComp panics.
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

func (c *Comp[T]) asTrack() Opt      { return comp.Track[T](&c.col) }
func (c *Comp[T]) asAdd() EditOpt    { return comp.Add[T](&c.col) }
func (c *Comp[T]) asLoad() CompToken { return LoadComp[T]() }

// Loadable is satisfied by *Comp[T] for any T — it lets LoadComps accept
// components (&comp) directly instead of naming their type again.
type Loadable interface {
	// asLoad is unexported so *Comp[T] is the only implementer — this is a
	// sealed interface, not an extension point.
	asLoad() CompToken
}

// Remove returns an EditOpt that removes component T from an entity.
func Remove[T any]() EditOpt { return comp.Del[T]() }

// LoadComp declares that a Load call may need to register component type
// T — see [ECS.Load]. The order tokens are passed to Load does not matter.
func LoadComp[T any]() CompToken { return reg.LoadComp[T]() }

// ProvidedComps collects LoadComps from every value that implements
// CompProvider, in order — values that don't implement it are skipped.
// Convenience for assembling an ECS.Load call from a mix of systems, e.g.
// ProvidedComps(movementSystem, collisionSystem).
func ProvidedComps(values ...any) []CompToken { return reg.ProvidedComps(values...) }

// LoadComps builds a CompToken per component, taking already-declared
// Comp[T] fields directly (&comp) instead of naming each type again —
// e.g. LoadComps(&m.pos, &m.vel) inside a CompProvider.LoadComps method.
func LoadComps(comps ...Loadable) []CompToken {
	out := make([]CompToken, len(comps))
	for i, c := range comps {
		out[i] = c.asLoad()
	}
	return out
}
