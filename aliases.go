package goke

import (
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/orch"
	"github.com/kjkrol/goke/internal/query"
	"github.com/kjkrol/goke/internal/reg"
)

type (
	// View matches entities by component mask and provides chunk-level and per-entity iteration.
	// Call All() or Filter() to set the iteration mode, then loop with Next().
	View = query.View

	// Col is a typed column handle for a tracked component.
	// Declare one per tracked component in each system struct, pass col.Track() to NewView,
	// then use col.Slice(v) in All-mode and col.At(v) in Filter-mode.
	Col[T any] = query.Col[T]

	// CompID is the unique integer identifier for a registered component type.
	CompID = comp.ID
	// CompMeta holds type metadata for a registered component (ID, size, alignment, reflect.Type).
	CompMeta = comp.Meta

	// Config holds initialization parameters for the ECS.
	Config = reg.Config
	// RunCtx provides methods to schedule systems sequentially or in parallel within a Plan.
	RunCtx = orch.RunCtx
	// Plan defines the execution order and concurrency of systems each tick.
	Plan = orch.Plan

	// Lookup provides a read-only view of component data for use inside system updates.
	Lookup = orch.Lookup
	// CmdBuf queues structural changes (add/remove component, destroy entity) during a tick.
	// Changes are applied at the next synchronization point, keeping iterators valid.
	CmdBuf = orch.CmdBuf

	// BlueprintOpt configures a View's entity filter by including or excluding component types.
	BlueprintOpt = comp.BlueprintOpt
)

// Include adds a required component type T to the View's filter.
// Only entities that possess this component will be matched.
func Include[T any]() BlueprintOpt { return comp.Include[T]() }

// Exclude adds an exclusion for component type T to the View's filter.
// Entities that possess this component will not be matched.
func Exclude[T any]() BlueprintOpt { return comp.Exclude[T]() }
