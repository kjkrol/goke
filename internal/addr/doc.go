// Package addr manages entity identifiers and their storage addresses.
//
// # Entry
//
// [Entry] is the full storage address of an entity:
//
//   - ArchID    — the archetype the entity belongs to
//   - ChunkPtr  — base address of the chunk holding the entity (stable across SwapChunks)
//   - Slot      — the entity's slot within that chunk
//   - Gen       — guards against stale access after an ID is recycled
//
// # Index
//
// [Index] is a flat slice keyed by the numeric index extracted from a [uid.UID64].
// It resolves an entity ID to its [Entry] in O(1) — no hash map, no scanning.
//
//	uid.UID64 → Unpack() → (index, generation)
//	                            │
//	                       Index[index]
//	                            │
//	                   Entry { ArchID, ChunkPtr, Slot, Gen }
//
// On every read the stored Generation is compared against the requested one.
// A mismatch means the slot was recycled and the lookup returns false.
// The slice grows on demand; removal marks the slot with [arch.NullID].
//
// [Index.GetUnchecked] skips both the bounds check and the generation
// check — for callers that already know the entity is alive (e.g. a prior
// [Index.Get] on the same ID) and want to avoid paying for the check twice.
//
// # Book
//
// [Book] is the address book: it combines entity ID lifecycle (uid pool) with
// the [Index] into a single owner.
//
//   - [Book.Seed]          — allocates a batch of IDs and registers their initial addresses
//   - [Book.Get]           — looks up the Entry for a given ID
//   - [Book.Move]          — updates the stored address after archetype migration
//   - [Book.MoveUnchecked] — same, without capacity/generation checks (bulk hot path)
//   - [Book.Delete]        — clears the address entry and recycles the ID
//   - [Book.PoolState]        — snapshots the ID pool's bookkeeping as plain data
//   - [Book.RestorePoolState] — replaces the ID pool's bookkeeping with a snapshot
//   - [Book.RestoreKnown]     — registers a previously-issued id at a given address,
//     bypassing pool allocation; fails if the id isn't recognized by the current
//     pool bookkeeping (e.g. after RestorePoolState)
//
// [Book.Index] is exported as a standalone read-only address index,
// separate from the ID pool.
package addr
