// Package addr manages entity identifiers and their storage addresses.
//
// # Entry
//
// [Entry] is the full storage address of an entity:
//
//   - ArchId      — the archetype the entity belongs to
//   - Pos         — the [colstore.Pos] (ChunkIdx + ChunkSlot) within that archetype's table
//   - Generation  — guards against stale access after an ID is recycled
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
//	                   Entry { ArchId, Pos, Generation }
//
// On every read the stored Generation is compared against the requested one.
// A mismatch means the slot was recycled and the lookup returns false.
// The slice grows on demand; removal marks the slot with [arch.NullID].
//
// # Book
//
// [Book] is the address book: it combines entity ID lifecycle (uid pool) with
// the [Index] into a single owner.
//
//   - [Book.Seed]   — allocates a batch of IDs and registers their initial addresses
//   - [Book.Get]    — looks up the Entry for a given ID
//   - [Book.Move]   — updates the stored address after archetype migration
//   - [Book.Delete] — clears the address entry and recycles the ID
//
// [Book.Index] is exported so that higher layers can hold a [*Index] for
// read-only lookups without access to the pool.
package addr
