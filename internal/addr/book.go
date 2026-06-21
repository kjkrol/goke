package addr

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
)

// Book is the address book: it combines entity ID lifecycle (uid pool) with
// the address [Index] into a single owner.
//
// [Book.Index] is exported so that the query layer can hold a [*Index] pointer
// for read-only entity lookups without access to the pool.
type Book struct {
	pool  uid.UID64Pool
	Index Index
}

func (b *Book) Init(cap int, freeCap int) {
	b.pool.Init(cap, freeCap)
	b.Index.Init(cap)
}

func (b *Book) Reset() {
	b.pool.Reset()
	b.Index.Reset()
}

// Seed allocates len(dst) entity IDs from the pool and registers their initial
// addresses in the Index. Intended for use as a colstore IDSeeder callback.
func (b *Book) Seed(dst []uid.UID64, archID arch.ID, pos colstore.Pos) {
	b.pool.NextN(dst)
	b.Index.EnsureCap(b.pool.PeekNextIndex())
	for i, id := range dst {
		b.Index.UpsertUnchecked(id, archID, colstore.Pos{Idx: pos.Idx, Slot: pos.Slot + colstore.Slot(i)})
	}
}

// Get looks up the Entry for the given entity ID.
// Returns false if the ID is invalid (wrong generation) or has no registered address.
func (b *Book) Get(id uid.UID64) (Entry, bool) {
	return b.Index.Get(id)
}

// Move updates the stored address for the given entity ID.
// Called after archetype migration to record the entity's new position.
func (b *Book) Move(id uid.UID64, archID arch.ID, pos colstore.Pos) {
	b.Index.Upsert(id, archID, pos)
}

// Delete clears the entity's address entry and recycles its ID back to the pool.
func (b *Book) Delete(id uid.UID64) {
	b.Index.Clear(id)
	b.pool.Release(id)
}
