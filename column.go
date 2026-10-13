package goke

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
)

// Column is Comp for a component type known only at run time: declare one with NewColumn and
// pass it where a &Comp[T] goes (NewQueryBuilder, NewFactory, NewEditorBuilder). Read its
// values through Base and At, which hand out the memory Comp.Slice and Comp.At do.
type Column struct {
	Idx  int
	typ  reflect.Type
	size uintptr
}

// NewColumn is the column of the component type t.
func NewColumn(t reflect.Type) *Column { return &Column{typ: t, size: t.Size()} }

// Type is the component type the column is of.
func (c *Column) Type() reflect.Type { return c.typ }

// Size is the size of one value, 0 for a tag.
func (c *Column) Size() uintptr { return c.size }

// Base is the column's array over the current All-mode chunk or Factory batch, len(cur.IDs)
// values from it; nil for a tag.
func (c *Column) Base(cur *Cursor) unsafe.Pointer {
	if c.Idx < 0 {
		return nil
	}
	return unsafe.Add(cur.Base, cur.Offsets[c.Idx])
}

// At is the value for the current Pick/Seek-mode entity; nil for a tag.
func (c *Column) At(cur *Cursor) unsafe.Pointer {
	if c.Idx < 0 {
		return nil
	}
	return unsafe.Add(cur.Base, cur.Offsets[c.Idx]+cur.Slot*c.size)
}

func (c *Column) asTrack() Opt      { return comp.TrackOf(c.typ, &c.Idx) }
func (c *Column) asAdd() EditOpt    { return comp.AddOf(c.typ, &c.Idx) }
func (c *Column) asLoad() CompToken { return LoadCompOf(c.typ) }

// OptColumn is OptComp for a component type known only at run time: declare one with
// NewOptColumn and pass it to QueryBuilder.Optional.
type OptColumn struct {
	Idx  int
	typ  reflect.Type
	size uintptr
}

// NewOptColumn is the optional column of the component type t.
func NewOptColumn(t reflect.Type) *OptColumn { return &OptColumn{typ: t, size: t.Size()} }

// Type is the component type the column is of.
func (c *OptColumn) Type() reflect.Type { return c.typ }

// Size is the size of one value, 0 for a tag.
func (c *OptColumn) Size() uintptr { return c.size }

// Present reports whether the component exists on the current chunk or entity.
func (c *OptColumn) Present(cur *Cursor) bool { return cur.OptPresent[c.Idx] }

// Base is the column's array over the current All-mode chunk, len(cur.IDs) values from it;
// nil where the chunk carries none.
func (c *OptColumn) Base(cur *Cursor) unsafe.Pointer {
	if !cur.OptPresent[c.Idx] {
		return nil
	}
	return unsafe.Add(cur.Base, cur.OptOffsets[c.Idx])
}

// At is the value for the current Pick/Seek-mode entity, nil where it carries none.
func (c *OptColumn) At(cur *Cursor) unsafe.Pointer {
	if !cur.OptPresent[c.Idx] {
		return nil
	}
	return unsafe.Add(cur.Base, cur.OptOffsets[c.Idx]+cur.Slot*c.size)
}

func (c *OptColumn) asOptTrack() Opt { return comp.OptionalOf(c.typ, &c.Idx) }

var (
	_ Trackable    = (*Column)(nil)
	_ Addable      = (*Column)(nil)
	_ Loadable     = (*Column)(nil)
	_ OptTrackable = (*OptColumn)(nil)
	_ iter.Cursor
)
