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
	// View matches entities by component mask and provides chunk-level and per-entity iteration.
	// Call All() or Filter() to set the iteration mode, then loop with Next().
	View = query.View

	// Col is a typed column handle for a tracked component.
	// Obtain one via Track[T](); use col.Slice(cursor) in All/Factory-mode
	// and col.At(cursor) in Filter-mode.
	Col[T any] = iter.Col[T]

	// Cursor holds the per-chunk or per-entity state populated by View.Next() and Factory.Next().
	// Pass view.Cursor or factory.Cursor to Col[T].Slice and Col[T].At.
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

	// Opt configures a View or Factory: tracks a data column, includes or excludes a component type.
	Opt = comp.BlueprintOpt

	// Factory bulk-spawns entities for a single archetype using a chunk-based iterator.
	// Call Create to set the count, then loop with Next; access entities via Entity and
	// components via col.Slice(&factory.Cursor).
	Factory = ent.Factory

	// Lookup provides cursor-based read access to a single entity's components.
	// Create once via CreateLookup; call Seek per entity access.
	Lookup = query.Lookup
)

// Track returns an Opt that registers T as a tracked data column and
// sets col.Idx when applied. Pass it to NewView or NewFactory.
func Track[T any](col *Col[T]) Opt { return comp.Track[T](col) }

// Include adds a required component type T to the View's filter.
// Only entities that possess this component will be matched.
func Include[T any]() Opt { return comp.Include[T]() }

// Exclude adds an exclusion for component type T to the View's filter.
// Entities that possess this component will not be matched.
func Exclude[T any]() Opt { return comp.Exclude[T]() }
