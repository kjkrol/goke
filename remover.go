package goke

import "github.com/kjkrol/uid"

// NewRemover creates a Remover ready for repeated CmdBufMassRemove calls.
// Unlike Migrator, Remover needs no configuration — it always removes the
// whole entity — so there is no builder, matching Factory's precedent of
// skipping an unnecessary builder step.
func (ecs *ECS) NewRemover() *Remover {
	return ecs.registry.CreateRemover()
}

// CmdBufMassRemove records a bulk removal of ids from a single source chunk.
// Call inside a system's Update after Query.All() + Next(): obtain snap from
// Query.ChunkSnapshot(), pass the ids you want to remove from cursor.IDs.
// Applied at the next Sync point: one Compact per chunk, one batch
// addr.Book update, zero per-entity archetype or pointer lookups beyond Slot.
func CmdBufMassRemove(cb *CmdBuf, remover *Remover, snap ChunkSnapshot, ids []uid.UID64) {
	cb.MassMigrate(remover, snap, ids)
}
