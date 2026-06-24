package goke

import (
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
	"github.com/kjkrol/goke/v2/internal/orch"
	"github.com/kjkrol/goke/v2/internal/query"
	"github.com/kjkrol/goke/v2/internal/reg"
	"github.com/kjkrol/goke/v2/iter"
)

type (
	// Query matches entities by component mask and provides three access
	// patterns: All (chunk-level iteration), Pick (per-entity iteration over a
	// given entity subset), and Seek (direct positioning on a single known
	// entity). Call All() or Pick() to set the iteration mode and loop with
	// Next(), or call Seek() directly for single-entity access.
	Query = query.Matcher

	// Cursor holds the per-chunk or per-entity state populated by Query.Next(),
	// Query.Seek(), and Factory.Next(). Pass query.Cursor or factory.Cursor
	// to Comp[T].Slice and Comp[T].At.
	Cursor = iter.Cursor

	// CompID is the unique integer identifier for a registered component type.
	CompID = comp.ID

	// Config holds initialization parameters for the ECS.
	Config = reg.Config
	// RunCtx provides methods to schedule systems sequentially or in parallel within a Plan.
	RunCtx = orch.RunCtx
	// Plan defines the execution order and concurrency of systems each tick.
	Plan = orch.Plan

	// CmdBuf queues structural changes (add/remove component, destroy entity) during a tick.
	// Changes are applied at the next synchronization point, keeping iterators valid.
	CmdBuf = orch.CmdBuf

	// Opt configures component access for a Query (Track / Include / Exclude).
	// It grants value access (read or write) within an entity's existing structure —
	// it never adds or removes components.
	Opt = comp.AccessOpt

	// Factory bulk-spawns entities for a single archetype using a chunk-based iterator.
	// Call Create to set the count, then loop with Next; access entities via Entity and
	// components via col.Slice(&factory.Cursor).
	Factory = ent.Factory

	// Editor applies add/remove component changes to an entity in a single
	// migration. Create once via NewEditorBuilder; call Update per entity, then
	// write added components' values via comp.At(&editor.Cursor).
	Editor = ent.Editor

	// EditOpt configures an Editor's structural change — Add or Del a component —
	// moving the entity to a different archetype. (Opt, by contrast, only accesses
	// components within an entity's existing structure.)
	EditOpt = comp.EditOpt
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

// Include adds a required component type T to the Query's filter.
// Only entities that possess this component will be matched.
func Include[T any]() Opt { return comp.Include[T]() }

// Exclude adds an exclusion for component type T to the Query's filter.
// Entities that possess this component will not be matched.
func Exclude[T any]() Opt { return comp.Exclude[T]() }

// Del returns an EditOpt that removes component T from an entity.
func Del[T any]() EditOpt { return comp.Del[T]() }
