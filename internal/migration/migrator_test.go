package migration_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
	"github.com/kjkrol/goke/v2/internal/migration"
	"github.com/kjkrol/goke/v2/iter"
)

type mPosition struct{ X, Y float64 }
type mVelocity struct{ VX, VY float64 }
type mTag struct{}

func newMgr() *ent.Manager {
	var m ent.Manager
	m.Init(ent.DefaultConfig(), nil)
	return &m
}

func internDefs(mi *comp.DefIndex) (pos, vel comp.Def) {
	pos = mi.Intern(reflect.TypeFor[mPosition]())
	vel = mi.Intern(reflect.TypeFor[mVelocity]())
	return
}

// spawnAll creates n entities using the given access spec and returns all IDs.
func spawnAll(m *ent.Manager, spec comp.AccessSpec, n int) []uid.UID64 {
	f := m.CreateFactory(spec)
	f.Create(n)
	var ids []uid.UID64
	for f.Next() {
		ids = append(ids, f.IDs...)
	}
	return ids
}

// applyByChunks groups consecutive ids by their current (ArchID, Ptr) and calls
// Migrate once per group — mirroring how ids arrive from Query.All()
// iteration in production, where each MassMigrate command covers one chunk.
func applyByChunks(m *ent.Manager, mig *migration.Migrator, ids []uid.UID64) {
	for start := 0; start < len(ids); {
		entry, ok := m.AddressBook.Get(ids[start])
		if !ok {
			start++
			continue
		}
		end := start + 1
		for end < len(ids) {
			next, ok := m.AddressBook.Get(ids[end])
			if !ok || next.ArchID != entry.ArchID || next.ChunkPtr != entry.ChunkPtr {
				break
			}
			end++
		}
		tbl := &m.ArchCatalog.Archetypes[entry.ArchID].Table
		snap := bulk.ChunkSnapshot{
			ArchID:   entry.ArchID,
			ChunkPtr: entry.ChunkPtr,
			ChunkIdx: tbl.ChunkIdxByPtr(entry.ChunkPtr),
		}
		mig.Migrate(snap, ids[start:end])
		start = end
	}
}

func TestMigrator_Migrate_Empty_IsNoOp(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 2)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	migrator.Migrate(bulk.ChunkSnapshot{}, nil)
	migrator.Migrate(bulk.ChunkSnapshot{}, []uid.UID64{})

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); !ok {
			t.Errorf("entity %v disappeared after no-op Migrate", id)
		}
	}
}

func TestMigrator_Migrate_AddComponent(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 3)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, ids)

	for _, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after Migrate", id)
		}
		if entry.ArchID == srcArchID {
			t.Errorf("entity %v still in src archetype", id)
		}
		mask := m.ArchCatalog.Archetypes[entry.ArchID].Mask()
		if !mask.IsSet(posDef.ID) {
			t.Errorf("entity %v: Pos missing after add-Vel migration", id)
		}
		if !mask.IsSet(velDef.ID) {
			t.Errorf("entity %v: Vel missing after add-Vel migration", id)
		}
	}
	if got := m.ArchCatalog.Archetypes[srcArchID].Table.Len(); got != 0 {
		t.Errorf("src Table.Len = %d after full batch migration; expected 0", got)
	}
}

func TestMigrator_Migrate_AddAlreadyPresentComponent_IsNoOp(t *testing.T) {
	// Migrator's AddDefs names a component every source entity already has:
	// dst resolves to the same archetype as src, so applyGroup must return
	// immediately — no copy, no compaction, no address-book update.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	_ = accessSpec.Comp(velDef)
	ids := spawnAll(m, accessSpec, 3) // already {Pos, Vel}

	entry0Before, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0Before.ArchID
	want := mPosition{X: 3.14, Y: 2.71}
	*(*mPosition)(m.ArchCatalog.Archetypes[srcArchID].Table.ComponentAt(entry0Before.ChunkPtr, entry0Before.Slot, posDef.ID)) = want

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity]))) // Vel already present on every source entity
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, ids)

	entry0After, ok := m.AddressBook.Get(ids[0])
	if !ok {
		t.Fatalf("entity %v missing after no-op migration", ids[0])
	}
	if entry0After.ArchID != srcArchID {
		t.Errorf("ArchID changed from %d to %d; expected untouched", srcArchID, entry0After.ArchID)
	}
	if entry0After.ChunkPtr != entry0Before.ChunkPtr || entry0After.Slot != entry0Before.Slot {
		t.Errorf("position moved from (%v,%d) to (%v,%d); expected untouched",
			entry0Before.ChunkPtr, entry0Before.Slot, entry0After.ChunkPtr, entry0After.Slot)
	}
	for _, id := range ids[1:] {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after no-op migration", id)
		}
		if entry.ArchID != srcArchID {
			t.Errorf("entity %v: ArchID changed from %d to %d; expected untouched", id, srcArchID, entry.ArchID)
		}
	}

	got := *(*mPosition)(m.ArchCatalog.Archetypes[srcArchID].Table.ComponentAt(entry0Before.ChunkPtr, entry0Before.Slot, posDef.ID))
	if got != want {
		t.Errorf("Pos after no-op migration: got %+v, want %+v", got, want)
	}
	if got := m.ArchCatalog.Archetypes[srcArchID].Table.Len(); got != 3 {
		t.Errorf("src Table.Len = %d after no-op migration; expected 3 (unchanged)", got)
	}
}

func TestMigrator_Migrate_RemoveComponent(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	_ = accessSpec.Comp(velDef)
	ids := spawnAll(m, accessSpec, 2)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Del[mVelocity]())
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, ids)

	for _, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after Migrate", id)
		}
		mask := m.ArchCatalog.Archetypes[entry.ArchID].Mask()
		if !mask.IsSet(posDef.ID) {
			t.Errorf("entity %v: Pos incorrectly removed", id)
		}
		if mask.IsSet(velDef.ID) {
			t.Errorf("entity %v: Vel still present after removal", id)
		}
	}
	if got := m.ArchCatalog.Archetypes[srcArchID].Table.Len(); got != 0 {
		t.Errorf("src Table.Len = %d after full migration; expected 0", got)
	}
}

func TestMigrator_Migrate_Unlink_RemovesEntitiesFromBook(t *testing.T) {
	// Removing the only component must delete the entities from the address book.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 3)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Del[mPosition]())
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, ids)

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("entity %v should be unlinked but still exists in addr.Book", id)
		}
	}
}

func TestMigrator_Migrate_AddAndRemove(t *testing.T) {
	// Add Vel, remove Pos → entities end up in a {Vel}-only archetype.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 2)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])), comp.Del[mPosition]())
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, ids)

	for _, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after add+remove migration", id)
		}
		mask := m.ArchCatalog.Archetypes[entry.ArchID].Mask()
		if mask.IsSet(posDef.ID) {
			t.Errorf("entity %v: Pos should have been removed", id)
		}
		if !mask.IsSet(velDef.ID) {
			t.Errorf("entity %v: Vel should have been added", id)
		}
	}
}

func TestMigrator_Migrate_DataPreserved(t *testing.T) {
	// Component values present before migration must be readable at the new position.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 1)

	entry0, _ := m.AddressBook.Get(ids[0])
	want := mPosition{X: 3.14, Y: 2.71}
	*(*mPosition)(m.ArchCatalog.Archetypes[entry0.ArchID].Table.ComponentAt(entry0.ChunkPtr, entry0.Slot, posDef.ID)) = want

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, ids)

	entryNew, _ := m.AddressBook.Get(ids[0])
	got := *(*mPosition)(m.ArchCatalog.Archetypes[entryNew.ArchID].Table.ComponentAt(entryNew.ChunkPtr, entryNew.Slot, posDef.ID))
	if got != want {
		t.Errorf("Pos after migration: got %+v, want %+v", got, want)
	}
}

func TestMigrator_Migrate_SrcCompaction_RemainingEntitiesConsistent(t *testing.T) {
	// Spawn 5 entities, migrate 3 (at slots 0, 2, 4).
	// The 2 remaining entities' addr.Book entries must match the compacted table state.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 5) // ids[0..4] seeded at slots 0..4

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, []uid.UID64{ids[0], ids[2], ids[4]})

	if got := m.ArchCatalog.Archetypes[srcArchID].Table.Len(); got != 2 {
		t.Errorf("src Table.Len = %d; expected 2", got)
	}

	// ids[1] stays in srcArch; its addr.Book entry must still be valid.
	entryE2, ok := m.AddressBook.Get(ids[1])
	if !ok {
		t.Fatal("ids[1] missing from addr.Book after compaction")
	}
	if entryE2.ArchID != srcArchID {
		t.Errorf("ids[1]: ArchID = %d; expected srcArch %d", entryE2.ArchID, srcArchID)
	}

	// ids[3] was moved to fill a hole; verify it's still accessible in srcArch.
	entryE4, ok2 := m.AddressBook.Get(ids[3])
	if !ok2 {
		t.Fatal("ids[3] missing from addr.Book after compaction")
	}
	if entryE4.ArchID != srcArchID {
		t.Errorf("ids[3]: ArchID = %d; expected srcArch %d", entryE4.ArchID, srcArchID)
	}
	// The two surviving entities must be at distinct positions.
	if entryE2.ChunkPtr == entryE4.ChunkPtr && entryE2.Slot == entryE4.Slot {
		t.Errorf("ids[1] and ids[3] share the same position ptr=%v slot=%d after compaction", entryE2.ChunkPtr, entryE2.Slot)
	}
}

func TestMigrator_Migrate_TwoBatches_SameDstArch(t *testing.T) {
	// Two sequential Migrate calls for the same source archetype must land
	// entities in the same destination archetype (dst is precomputed per srcArch).
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)

	ids1 := spawnAll(m, accessSpec, 1)
	ids2 := spawnAll(m, accessSpec, 1)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunks(m, migrator, ids1)
	applyByChunks(m, migrator, ids2)

	entry1, _ := m.AddressBook.Get(ids1[0])
	entry2, _ := m.AddressBook.Get(ids2[0])
	if entry1.ArchID != entry2.ArchID {
		t.Errorf("expected same dstArch for both batches; got %d and %d", entry1.ArchID, entry2.ArchID)
	}
}

func TestMigrator_Migrate_SlotAlignedFastPath(t *testing.T) {
	// TableVer match + SlotAligned: slots are synthesized (0..n-1) with zero addr.Book
	// reads. Entities must land in the destination archetype with data intact.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 4)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID
	srcTable := &m.ArchCatalog.Archetypes[srcArchID].Table

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	snap := bulk.ChunkSnapshot{
		ArchID:      srcArchID,
		ChunkPtr:    entry0.ChunkPtr,
		ChunkIdx:    srcTable.ChunkIdxByPtr(entry0.ChunkPtr),
		TableVer:    srcTable.Version(),
		SlotAligned: true,
	}
	migrator.Migrate(snap, ids)

	for _, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after prefix fast-path migration", id)
		}
		if !m.ArchCatalog.Archetypes[entry.ArchID].Mask().IsSet(velDef.ID) {
			t.Errorf("entity %v: Vel missing after migration", id)
		}
	}
	if got := srcTable.Len(); got != 0 {
		t.Errorf("src Table.Len = %d; expected 0", got)
	}
}

func TestMigrator_Migrate_StaleVersion_RevalidatesEntities(t *testing.T) {
	// A structural change between capture and apply (version bump) forces the
	// slow path: dead entities are skipped, survivors migrate from their
	// current (possibly relocated) positions.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)
	_ = velDef

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 4)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID
	srcTable := &m.ArchCatalog.Archetypes[srcArchID].Table

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	// Capture snap, then mutate the table: remove ids[1] (bumps version and
	// relocates the tail entity into its slot).
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

	migrator.Migrate(snap, ids)

	if _, ok := m.AddressBook.Get(ids[1]); ok {
		t.Error("ids[1] was removed before apply; must not resurface")
	}
	for _, id := range []uid.UID64{ids[0], ids[2], ids[3]} {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after stale-version migration", id)
		}
		if entry.ArchID == srcArchID {
			t.Errorf("entity %v still in src archetype", id)
		}
	}
	if got := srcTable.Len(); got != 0 {
		t.Errorf("src Table.Len = %d; expected 0", got)
	}
}

func TestMigrator_Migrate_MultiSrcArch_PerChunkCalls(t *testing.T) {
	// Entities from two different source archetypes migrate correctly when
	// each chunk is applied by its own Migrate call — as the CmdBuf does
	// with one MassMigrate command per Query.All() chunk.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	// Arch X: {Pos}
	var specPos comp.AccessSpec
	_ = specPos.Comp(posDef)
	idsPos := spawnAll(m, specPos, 3) // srcArch X

	// Arch Y: {Pos, Vel}
	var specPosVel comp.AccessSpec
	_ = specPosVel.Comp(posDef)
	_ = specPosVel.Comp(velDef)
	idsPosVel := spawnAll(m, specPosVel, 2) // srcArch Y

	entryX, _ := m.AddressBook.Get(idsPos[0])
	entryY, _ := m.AddressBook.Get(idsPosVel[0])
	srcArchX := entryX.ArchID
	srcArchY := entryY.ArchID

	// Migrator adds mTag to every entity regardless of source archetype.
	tagCol := new(iter.ArrayRef[mTag])
	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add[mTag](tagCol))
	tagDef := editSpec.AddDefs[0]
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	// Interleaved ids from both archetypes; applyByChunks splits them into
	// per-(arch, chunk) Migrate calls.
	mixed := []uid.UID64{idsPos[0], idsPosVel[0], idsPos[1], idsPosVel[1], idsPos[2]}
	applyByChunks(m, migrator, mixed)

	// All 5 entities must be in new archetypes (not srcArchX or srcArchY).
	for _, id := range idsPos {
		e, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("idsPos entity %v missing after migration", id)
		}
		if e.ArchID == srcArchX {
			t.Errorf("idsPos entity %v still in srcArchX", id)
		}
		if !m.ArchCatalog.Archetypes[e.ArchID].Mask().IsSet(tagDef.ID) {
			t.Errorf("idsPos entity %v: Tag missing after migration", id)
		}
	}
	for _, id := range idsPosVel {
		e, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("idsPosVel entity %v missing after migration", id)
		}
		if e.ArchID == srcArchY {
			t.Errorf("idsPosVel entity %v still in srcArchY", id)
		}
		if !m.ArchCatalog.Archetypes[e.ArchID].Mask().IsSet(tagDef.ID) {
			t.Errorf("idsPosVel entity %v: Tag missing after migration", id)
		}
	}
	// Both source archetypes must be empty.
	if got := m.ArchCatalog.Archetypes[srcArchX].Table.Len(); got != 0 {
		t.Errorf("srcArchX Table.Len = %d; expected 0", got)
	}
	if got := m.ArchCatalog.Archetypes[srcArchY].Table.Len(); got != 0 {
		t.Errorf("srcArchY Table.Len = %d; expected 0", got)
	}
}

func TestMigrator_LazyResolve_NoArchetypeExplosion(t *testing.T) {
	// Two coexisting migrators whose adds are each other's dels must not
	// materialize intermediate archetypes: destinations are resolved lazily,
	// straight by target mask, only for source archetypes actually migrated.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 4)

	var fwdSpec, revSpec comp.EditSpec
	fwdSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])), comp.Add(new(iter.ArrayRef[mTag])), comp.Del[mPosition]())
	revSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mPosition])), comp.Del[mVelocity](), comp.Del[mTag]())

	before := m.ArchCatalog.Len()
	fwd := migration.New(&m.AddressBook, &m.ArchCatalog, fwdSpec)
	rev := migration.New(&m.AddressBook, &m.ArchCatalog, revSpec)
	if got := m.ArchCatalog.Len(); got != before {
		t.Fatalf("constructing migrators created %d archetypes; expected none", got-before)
	}

	applyByChunks(m, fwd, ids)
	applyByChunks(m, rev, ids)
	applyByChunks(m, fwd, ids)

	if got := m.ArchCatalog.Len() - before; got != 1 {
		t.Errorf("expected exactly 1 new archetype (the fwd target), got %d", got)
	}
	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); !ok {
			t.Fatalf("entity %v lost during fwd/rev/fwd cycle", id)
		}
	}
}

func TestMigrator_SourceArchetypeCreatedAfterConstruction(t *testing.T) {
	// A Migrator has no registered set of source archetypes: dst is resolved
	// lazily per source, so an archetype born after the Migrator was built
	// migrates just like a pre-existing one.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	migrator := migration.New(&m.AddressBook, &m.ArchCatalog, editSpec)

	// The source archetype {Pos} comes into existence only now.
	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 3)

	applyByChunks(m, migrator, ids)

	for _, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after migration from a late-born archetype", id)
		}
		mask := m.ArchCatalog.Archetypes[entry.ArchID].Mask()
		if !mask.IsSet(velDef.ID) || !mask.IsSet(posDef.ID) {
			t.Errorf("entity %v: expected {Pos, Vel} archetype after migration", id)
		}
	}
}
