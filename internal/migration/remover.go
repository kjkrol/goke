package migration

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/addr"
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/colstore"
)

// Remover removes whole entities in bulk from a batch sharing the same
// source archetype. Unlike Migrator it needs no configuration: every call
// removes every given entity outright, compacting the source table in one
// batched pass via colstore.Defragmenter instead of a per-entity swap-and-pop.
type Remover struct {
	addrBook    *addr.Book
	archCatalog *arch.Catalog

	defrag colstore.Defragmenter // by value: shares Remover's heap allocation

	// scratch is reused across Migrate calls; grows lazily, never shrinks.
	scratch struct {
		ids      []uid.UID64
		slotRefs []colstore.SlotRef
	}
}

func NewRemover(book *addr.Book, catalog *arch.Catalog) *Remover {
	return &Remover{addrBook: book, archCatalog: catalog}
}

// Migrate satisfies bulk.Migrator: it removes ids outright rather than
// migrating them to another archetype.
func (r *Remover) Migrate(snap bulk.ChunkSnapshot, ids []uid.UID64) {
	if len(ids) == 0 {
		return
	}
	srcTable := &r.archCatalog.Archetypes[snap.ArchID].Table
	validIDs, slotRefs := resolveSlotRefs(r.addrBook, srcTable, snap, ids, &r.scratch.ids, &r.scratch.slotRefs, nil)
	if len(validIDs) == 0 {
		return
	}
	removeBatch(r.addrBook, &r.defrag, snap.ArchID, srcTable, validIDs, slotRefs)
}
