package addr

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/colstore"
)

// Book is the address book: entity ID lifecycle (uid pool) plus the address
// [Index] under a single owner. [Book.Index] is exported so the query layer
// can hold a read-only [*Index] without access to the pool.
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

// Seed allocates len(dst) entity IDs from the pool and registers their
// addresses: consecutive slots in ptr's chunk starting at startSlot.
func (b *Book) Seed(dst []uid.UID64, archID arch.ID, ptr unsafe.Pointer, startSlot colstore.Slot) {
	b.pool.NextN(dst)
	b.Index.EnsureCap(b.pool.PeekNextIndex())
	for i, id := range dst {
		b.Index.UpsertUnchecked(id, archID, ptr, startSlot+colstore.Slot(i))
	}
}

func (b *Book) Get(id uid.UID64) (Entry, bool) {
	return b.Index.Get(id)
}

func (b *Book) Move(id uid.UID64, archID arch.ID, ptr unsafe.Pointer, slot colstore.Slot) {
	b.Index.Upsert(id, archID, ptr, slot)
}

// MoveUnchecked is Move minus the capacity and generation checks; id must
// already exist in the Index. Bulk-migration hot path.
func (b *Book) MoveUnchecked(id uid.UID64, archID arch.ID, ptr unsafe.Pointer, slot colstore.Slot) {
	b.Index.UpsertUnchecked(id, archID, ptr, slot)
}

// Delete clears the entity's address entry and recycles its ID back to the pool.
func (b *Book) Delete(id uid.UID64) {
	b.Index.Clear(id)
	b.pool.Release(id)
}
