package migration_test

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/migration"
)

func TestRemover_Migrate_Empty_IsNoOp(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 2)

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)
	remover.Migrate(bulk.ChunkSnapshot{}, nil)
	remover.Migrate(bulk.ChunkSnapshot{}, []uid.UID64{})

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); !ok {
			t.Errorf("entity %v disappeared after no-op Migrate", id)
		}
	}
}

func TestRemover_Migrate_RemovesEntitiesFromBook(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 3)

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)
	applyByChunks(m, remover, ids)

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("entity %v should be removed but still exists in addr.Book", id)
		}
	}
}

func TestRemover_Migrate_TableEmptiedAfterFullRemoval(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 4)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)
	applyByChunks(m, remover, ids)

	if got := m.ArchCatalog.Archetypes[srcArchID].Table.Len(); got != 0 {
		t.Errorf("src Table.Len = %d after full removal; expected 0", got)
	}
}

func TestRemover_Migrate_PartialRemoval_SurvivorsConsistent(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 5) // ids[0..4] seeded at slots 0..4

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)
	applyByChunks(m, remover, []uid.UID64{ids[0], ids[2], ids[4]})

	if got := m.ArchCatalog.Archetypes[srcArchID].Table.Len(); got != 2 {
		t.Errorf("src Table.Len = %d; expected 2", got)
	}
	for _, id := range []uid.UID64{ids[0], ids[2], ids[4]} {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("entity %v should be removed but still exists in addr.Book", id)
		}
	}

	entry1, ok1 := m.AddressBook.Get(ids[1])
	if !ok1 {
		t.Fatal("ids[1] missing from addr.Book after partial removal")
	}
	if entry1.ArchID != srcArchID {
		t.Errorf("ids[1]: ArchID = %d; expected srcArch %d", entry1.ArchID, srcArchID)
	}
	entry3, ok3 := m.AddressBook.Get(ids[3])
	if !ok3 {
		t.Fatal("ids[3] missing from addr.Book after partial removal")
	}
	if entry3.ArchID != srcArchID {
		t.Errorf("ids[3]: ArchID = %d; expected srcArch %d", entry3.ArchID, srcArchID)
	}
	if entry1.ChunkPtr == entry3.ChunkPtr && entry1.Slot == entry3.Slot {
		t.Errorf("ids[1] and ids[3] share the same position ptr=%v slot=%d after compaction", entry1.ChunkPtr, entry1.Slot)
	}
}

func TestRemover_Migrate_SlotAlignedFastPath(t *testing.T) {
	// TableVer match + SlotAligned: slots are synthesized (0..n-1) with zero
	// addr.Book reads. Entities must be gone and survivors untouched.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 4)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID
	srcTable := &m.ArchCatalog.Archetypes[srcArchID].Table

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)

	snap := bulk.ChunkSnapshot{
		ArchID:      srcArchID,
		ChunkPtr:    entry0.ChunkPtr,
		ChunkIdx:    srcTable.ChunkIdxByPtr(entry0.ChunkPtr),
		TableVer:    srcTable.Version(),
		SlotAligned: true,
	}
	remover.Migrate(snap, ids)

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("entity %v should be removed but still exists in addr.Book", id)
		}
	}
	if got := srcTable.Len(); got != 0 {
		t.Errorf("src Table.Len = %d; expected 0", got)
	}
}

func TestRemover_Migrate_StaleVersion_RevalidatesEntities(t *testing.T) {
	// A structural change between capture and apply (version bump) forces the
	// slow path: dead entities are skipped, survivors are removed from their
	// current (possibly relocated) positions.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 4)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID
	srcTable := &m.ArchCatalog.Archetypes[srcArchID].Table

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)

	snap := bulk.ChunkSnapshot{
		ArchID:      srcArchID,
		ChunkPtr:    entry0.ChunkPtr,
		ChunkIdx:    srcTable.ChunkIdxByPtr(entry0.ChunkPtr),
		TableVer:    srcTable.Version(),
		SlotAligned: true,
	}
	if !m.Remove(ids[1]) {
		t.Fatal("failed to remove ids[1]")
	}

	remover.Migrate(snap, ids)

	if _, ok := m.AddressBook.Get(ids[1]); ok {
		t.Error("ids[1] was removed before apply; must not resurface")
	}
	for _, id := range []uid.UID64{ids[0], ids[2], ids[3]} {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("entity %v should be removed but still exists in addr.Book", id)
		}
	}
	if got := srcTable.Len(); got != 0 {
		t.Errorf("src Table.Len = %d; expected 0", got)
	}
}

func TestRemover_Migrate_StaleVersion_AllEntitiesDead_NoOp(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 3)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID
	srcTable := &m.ArchCatalog.Archetypes[srcArchID].Table

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)

	snap := bulk.ChunkSnapshot{
		ArchID:      srcArchID,
		ChunkPtr:    entry0.ChunkPtr,
		ChunkIdx:    srcTable.ChunkIdxByPtr(entry0.ChunkPtr),
		TableVer:    srcTable.Version(),
		SlotAligned: true,
	}
	for _, id := range ids {
		if !m.Remove(id) {
			t.Fatalf("failed to remove %v", id)
		}
	}

	remover.Migrate(snap, ids) // must not panic; nothing left to remove

	if got := srcTable.Len(); got != 0 {
		t.Errorf("src Table.Len = %d; expected 0", got)
	}
}

func TestRemover_Migrate_MultiSrcArch_PerChunkCalls(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var specPos comp.AccessSpec
	_ = specPos.Comp(posDef)
	idsPos := spawnAll(m, specPos, 3) // srcArch X

	var specPosVel comp.AccessSpec
	_ = specPosVel.Comp(posDef)
	_ = specPosVel.Comp(velDef)
	idsPosVel := spawnAll(m, specPosVel, 2) // srcArch Y

	entryX, _ := m.AddressBook.Get(idsPos[0])
	entryY, _ := m.AddressBook.Get(idsPosVel[0])
	srcArchX := entryX.ArchID
	srcArchY := entryY.ArchID

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)

	mixed := []uid.UID64{idsPos[0], idsPosVel[0], idsPos[1], idsPosVel[1], idsPos[2]}
	applyByChunks(m, remover, mixed)

	for _, id := range idsPos {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("idsPos entity %v should be removed but still exists in addr.Book", id)
		}
	}
	for _, id := range idsPosVel {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("idsPosVel entity %v should be removed but still exists in addr.Book", id)
		}
	}
	if got := m.ArchCatalog.Archetypes[srcArchX].Table.Len(); got != 0 {
		t.Errorf("srcArchX Table.Len = %d; expected 0", got)
	}
	if got := m.ArchCatalog.Archetypes[srcArchY].Table.Len(); got != 0 {
		t.Errorf("srcArchY Table.Len = %d; expected 0", got)
	}
}

func TestRemover_Migrate_IDRecycling(t *testing.T) {
	// Removed entities' slots must be reclaimed and reusable by later spawns —
	// checked by reachability only, not a specific reuse order.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 3)

	remover := migration.NewRemover(&m.AddressBook, &m.ArchCatalog)
	applyByChunks(m, remover, ids)

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Fatalf("entity %v should be removed but still exists in addr.Book", id)
		}
	}

	newIDs := spawnAll(m, accessSpec, 3)
	if len(newIDs) != 3 {
		t.Fatalf("expected 3 new entities, got %d", len(newIDs))
	}
	for _, id := range newIDs {
		if _, ok := m.AddressBook.Get(id); !ok {
			t.Errorf("newly spawned entity %v not addressable", id)
		}
	}
}
