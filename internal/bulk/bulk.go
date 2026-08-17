package bulk

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/colstore"
)

// ChunkSnapshot is the address of a whole chunk captured at a point in time —
// the ephemeral, chunk-level counterpart of the durable per-entity addr.Entry.
// Positions derived from it are valid only while TableVer still matches the
// source table's structural version (no removal or relocation since capture);
// that short lifetime is why it may carry ChunkIdx, which Entry omits.
type ChunkSnapshot struct {
	ArchID      arch.ID        // archetype owning the chunk
	ChunkPtr    unsafe.Pointer // chunk's base memory address — its stable identity
	ChunkIdx    colstore.Idx   // chunk's position within the table's chunk pack
	TableVer    uint32         // source table's structural version at capture time
	SlotAligned bool           // the accompanying ids[i] lives at slot i — no per-entity Slot lookups needed
}

// Migrator is satisfied by any type that can apply a bulk archetype
// migration to a batch of entity IDs from a single source chunk.
type Migrator interface {
	Migrate(snap ChunkSnapshot, ids []uid.UID64)
}

// ValueMigrator is satisfied by any type that can apply a bulk archetype
// migration to a batch of entity IDs from a single source chunk, writing
// payload (n elements of a caller-known type, one per id) into the newly
// added column as part of the same migration.
type ValueMigrator interface {
	MigrateWithValue(snap ChunkSnapshot, ids []uid.UID64, payload unsafe.Pointer)
}

// Spawner is satisfied by any type that can create count new entities of a
// fixed archetype in one call, returning their ids.
type Spawner interface {
	Spawn(count int) []uid.UID64
}
