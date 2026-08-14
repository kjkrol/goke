package goke

import (
	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
	"github.com/kjkrol/goke/v2/internal/migration"
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

	// Migrator applies a fixed set of structural changes to a batch of entities
	// sharing the same source archetype. Create once via NewMigratorBuilder;
	// call via CmdBufMassMigrate inside a system to defer bulk migration to Sync.
	Migrator = migration.Migrator

	// Remover bulk-removes whole entities from a batch sharing the same source
	// archetype, compacting the source table in one batched pass instead of a
	// per-entity swap-and-pop. Create once via NewRemover; call via
	// CmdBufMassRemove inside a system to defer bulk removal to Sync.
	Remover = migration.Remover

	// ValueMigrator is Migrator's value-carrying counterpart: it adds exactly
	// one component and lets the caller write a per-entity value for it,
	// computed at enqueue time and applied at Sync alongside the migration.
	// Create once via NewValueMigratorBuilder; call via CmdBufMassMigrateInit.
	ValueMigrator = migration.ValueMigrator

	// EditOpt configures an Editor's structural change — Add or Del a component —
	// moving the entity to a different archetype. (Opt, by contrast, only accesses
	// components within an entity's existing structure.)
	EditOpt = comp.EditOpt

	// ChunkSnapshot is a point-in-time capture of a single table chunk.
	// Obtained from Query.ChunkSnapshot() during Query.All() iteration; passed
	// to CmdBufMassMigrate so the Migrator can skip per-entity addr.Book lookups.
	ChunkSnapshot = bulk.ChunkSnapshot
)
