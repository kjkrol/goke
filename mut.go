package goke

import (
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
)

// Mut is Comp for writing: Slice and At also count a change of the cursor's chunk (Cursor.Ver,
// ChunkSnapshot.Ver); a write through Comp goes unseen but by WithMutCheck. The count is the
// chunk's, so parallel writers of one archetype race on it; take a chunk's Mut slice first, or
// the slices taken before it lose their bounds check elimination.
type Mut[T any] struct {
	col iter.ArrayRef[T]
}

// Slice is Comp.Slice, counting a change of the current All-mode chunk or Factory batch.
func (c *Mut[T]) Slice(cur *Cursor) []T {
	*cur.Ver++
	return c.col.Slice(cur)
}

// At is Comp.At, counting a change of the chunk of the current Pick/Seek-mode entity.
func (c *Mut[T]) At(cur *Cursor) *T {
	*cur.Ver++
	return c.col.At(cur)
}

// Index is the column's track position, as a Column reads it; set as the column is bound.
func (c *Mut[T]) Index() int { return c.col.Idx }

func (c *Mut[T]) asTrack() Opt      { return comp.Track[T](&c.col) }
func (c *Mut[T]) asAdd() EditOpt    { return comp.Add[T](&c.col) }
func (c *Mut[T]) asLoad() CompToken { return LoadComp[T]() }
