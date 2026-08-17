package goke

import (
	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/internal/orch"
	"github.com/kjkrol/goke/v3/internal/reg"
	"github.com/kjkrol/goke/v3/iter"
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

	// Opt configures component access for a Query (Track / Include / Exclude).
	// It grants value access (read or write) within an entity's existing structure —
	// it never adds or removes components.
	Opt = comp.AccessOpt

	// Factory bulk-spawns entities for a single archetype using a chunk-based iterator.
	// Call Create to set the count, then loop with Next; access entities via Entity and
	// components via col.Slice(&factory.Cursor).
	Factory = ent.Factory

	// Editor applies a fixed set of structural changes to a batch of entities
	// sharing the same source archetype. Create once via NewEditorBuilder;
	// call via cb.Migrate inside a system to defer bulk migration to Sync.
	Editor = ent.Editor

	// ValueEditor is Editor's value-carrying counterpart: it adds exactly
	// one component and lets the caller write a per-entity value for it,
	// computed at enqueue time and applied at Sync alongside the migration.
	// Create once via NewValueEditorBuilder; call via CmdBufAddCompValue.
	ValueEditor = ent.ValueEditor

	// EditOpt configures a structural change — Add or Remove a component —
	// used when building an Editor or ValueEditor. (Opt, by contrast, only
	// accesses components within an entity's existing structure.)
	EditOpt = comp.EditOpt

	// ChunkSnapshot is a point-in-time capture of a single table chunk.
	// Obtained from Query.ChunkSnapshot() during Query.All() iteration; passed
	// to cb.Migrate so the Editor can skip per-entity addr.Book lookups.
	ChunkSnapshot = bulk.ChunkSnapshot
)
