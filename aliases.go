package goke

import (
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
	"github.com/kjkrol/goke/internal/orch"
	"github.com/kjkrol/goke/internal/query"
	"github.com/kjkrol/goke/internal/reg"
	"github.com/kjkrol/goke/iter"
)

type (
	// Matcher matches entities by component mask and provides three access
	// patterns: All (chunk-level iteration), Pick (per-entity iteration over a
	// given entity subset), and Seek (direct positioning on a single known
	// entity). Call All() or Pick() to set the iteration mode and loop with
	// Next(), or call Seek() directly for single-entity access.
	Matcher = query.Matcher

	// Col is a typed column handle for a tracked component.
	// Obtain one via Track[T](); use col.Slice(cursor) in All/Factory-mode
	// and col.At(cursor) in Pick/Seek-mode.
	Col[T any] = iter.Col[T]

	// Cursor holds the per-chunk or per-entity state populated by Matcher.Next(),
	// Matcher.Seek(), and Factory.Next(). Pass matcher.Cursor or factory.Cursor
	// to Col[T].Slice and Col[T].At.
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

	// Opt configures component access for a Matcher (Track / Include / Exclude).
	// It grants value access (read or write) within an entity's existing structure —
	// it never adds or removes components.
	Opt = comp.AccessOpt

	// Factory bulk-spawns entities for a single archetype using a chunk-based iterator.
	// Call Create to set the count, then loop with Next; access entities via Entity and
	// components via col.Slice(&factory.Cursor).
	Factory = ent.Factory

	// Editor applies add/remove component changes to an entity in a single
	// migration. Create once via CreateEditor; call Update per entity, then write
	// added components' values via col.At(&editor.Cursor).
	Editor = ent.Editor

	// EditOpt configures an Editor's structural change — Add or Del a component —
	// moving the entity to a different archetype. (Opt, by contrast, only accesses
	// components within an entity's existing structure.)
	EditOpt = comp.EditOpt
)

// Track returns an Opt that registers T as a tracked data column and
// sets col.Idx when applied. Pass it to NewMatcher or NewFactory.
func Track[T any](col *Col[T]) Opt { return comp.Track[T](col) }

// Include adds a required component type T to the Matcher's filter.
// Only entities that possess this component will be matched.
func Include[T any]() Opt { return comp.Include[T]() }

// Exclude adds an exclusion for component type T to the Matcher's filter.
// Entities that possess this component will not be matched.
func Exclude[T any]() Opt { return comp.Exclude[T]() }

// Add returns an EditOpt that adds component T to an entity and binds col so its
// value can be written via col.At(&editor.Cursor) after Update.
func Add[T any](col *Col[T]) EditOpt { return comp.Add[T](col) }

// Del returns an EditOpt that removes component T from an entity.
func Del[T any]() EditOpt { return comp.Del[T]() }
