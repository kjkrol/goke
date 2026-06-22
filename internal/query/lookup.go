package query

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/addr"
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

// Lookup provides cursor-based read access to a single entity's components.
// Create once; call Seek per entity access.
type Lookup struct {
	Cursor      iter.Cursor
	addrIndex   *addr.Index
	archCatalog *arch.Catalog
	compIDs     []comp.ID
	offsets     [arch.MaxID][]uintptr
	table       *colstore.Table
	lastArchID  arch.ID
}

func (l *Lookup) Init(addrIndex *addr.Index, archCatalog *arch.Catalog, compIDs []comp.ID) {
	l.addrIndex = addrIndex
	l.archCatalog = archCatalog
	l.compIDs = compIDs
	l.lastArchID = arch.NullID
}

// Seek positions the Cursor at entID's storage slot.
// Returns false if the entity does not exist or has been recycled.
//
// Consecutive Seeks that resolve to the same archetype reuse the cached table
// and column offsets, updating only the cursor's chunk and slot. Looping Seek
// over a set of entities from one archetype therefore resolves the archetype
// just once.
func (l *Lookup) Seek(entID uid.UID64) bool {
	entry, ok := l.addrIndex.Get(entID)
	if !ok {
		return false
	}
	if entry.ArchId != l.lastArchID {
		l.table = &l.archCatalog.Archetypes[entry.ArchId].Table
		offs := l.offsets[entry.ArchId]
		if offs == nil {
			offs = l.table.BakeOffsets(l.compIDs)
			l.offsets[entry.ArchId] = offs
		}
		l.Cursor.Offsets = offs
		l.lastArchID = entry.ArchId
	}
	l.table.PointCursor(&l.Cursor, entry.Pos)
	return true
}
