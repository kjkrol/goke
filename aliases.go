package goke

import (
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
	"github.com/kjkrol/goke/v2/internal/orch"
	"github.com/kjkrol/goke/v2/internal/reg"
	"github.com/kjkrol/goke/v2/iter"
)

type (
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
