package goke

import "github.com/kjkrol/uid"

// MigratorBuilder assembles a Migrator's structural-change options. Start with
// NewMigratorBuilder, optionally chain Delete, and finish with Build.
type MigratorBuilder struct {
	ecs  *ECS
	opts []EditOpt
}

// NewMigratorBuilder starts a MigratorBuilder, adding the given components
// (equivalent to Add[T] for each).
func (ecs *ECS) NewMigratorBuilder(comps ...Addable) *MigratorBuilder {
	b := &MigratorBuilder{ecs: ecs, opts: make([]EditOpt, 0, len(comps))}
	for _, c := range comps {
		b.opts = append(b.opts, c.asAdd())
	}
	return b
}

// Delete adds component types to remove, built via Del[T]().
func (b *MigratorBuilder) Delete(opts ...EditOpt) *MigratorBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates the Migrator from the accumulated options.
func (b *MigratorBuilder) Build() *Migrator {
	return b.ecs.registry.CreateMigrator(b.opts...)
}

// CmdBufMassMigrate records a bulk migration of ids from a single source chunk.
// Call inside a system's Update after Query.All() + Next(): obtain snap from
// Query.ChunkSnapshot(), pass the ids you want to migrate from cursor.IDs.
// Applied at the next Sync point: one CompactHoles per chunk, one batch
// addr.Book update, zero per-entity archetype or pointer lookups beyond Slot.
func CmdBufMassMigrate(cb *CmdBuf, migrator *Migrator, snap ChunkSnapshot, ids []uid.UID64) {
	cb.MassMigrate(migrator, snap, ids)
}
