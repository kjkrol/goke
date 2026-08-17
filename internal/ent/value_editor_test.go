package ent_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/iter"
)

// applyByChunksWithValue mirrors applyByChunks (editor_test.go) but for
// ValueEditor: groups consecutive ids by their current (ArchID, Ptr) and
// calls MigrateWithValue once per group, slicing values in lockstep — one
// call per chunk, exactly as CmdBufAddCompValue's production callers do.
func applyByChunksWithValue[T any](m *ent.Manager, vm *ent.ValueEditor, ids []uid.UID64, values []T) {
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
		payload := unsafe.Pointer(unsafe.SliceData(values[start:end]))
		vm.MigrateWithValue(snap, ids[start:end], payload)
		start = end
	}
}

func readVel(m *ent.Manager, velDef comp.Def, id uid.UID64) mVelocity {
	entry, ok := m.AddressBook.Get(id)
	if !ok {
		panic("readVel: entity not found")
	}
	tbl := &m.ArchCatalog.Archetypes[entry.ArchID].Table
	ptr := tbl.ComponentAt(entry.ChunkPtr, entry.Slot, velDef.ID)
	if ptr == nil {
		panic("readVel: entity's archetype has no velocity column")
	}
	return *(*mVelocity)(ptr)
}

func TestValueEditor_MigrateWithValue_Empty_IsNoOp(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 2)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	vm.MigrateWithValue(bulk.ChunkSnapshot{}, nil, nil)
	vm.MigrateWithValue(bulk.ChunkSnapshot{}, []uid.UID64{}, nil)

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); !ok {
			t.Errorf("entity %v disappeared after no-op MigrateWithValue", id)
		}
	}
}

func TestValueEditor_MigrateWithValue_ValuesWrittenPerEntity(t *testing.T) {
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
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	values := []mVelocity{{VX: 10, VY: 10}, {VX: 20, VY: 20}, {VX: 30, VY: 30}}
	applyByChunksWithValue(m, vm, ids, values)

	for i, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after MigrateWithValue", id)
		}
		if entry.ArchID == srcArchID {
			t.Errorf("entity %v: still in source archetype", id)
		}
		if got := readVel(m, velDef, id); got != values[i] {
			t.Errorf("entity %v: velocity = %+v; want %+v", id, got, values[i])
		}
	}
}

func TestValueEditor_MigrateWithValue_StaleVersion_ValuesStayAlignedAfterDrop(t *testing.T) {
	// The core case this whole mechanism exists for: a structural change
	// between capture and apply forces the slow path, which drops a dead
	// id — survivors' values must still match their own original index, not
	// shift into a neighbor's slot.
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
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

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

	values := []mVelocity{{VX: 0, VY: 0}, {VX: 999, VY: 999}, {VX: 2, VY: 2}, {VX: 3, VY: 3}}
	payload := unsafe.Pointer(unsafe.SliceData(values))
	vm.MigrateWithValue(snap, ids, payload)

	if _, ok := m.AddressBook.Get(ids[1]); ok {
		t.Error("ids[1] was removed before apply; must not resurface")
	}
	for _, i := range []int{0, 2, 3} {
		if got := readVel(m, velDef, ids[i]); got != values[i] {
			t.Errorf("entity ids[%d]: velocity = %+v; want %+v (values misaligned after slow-path drop)", i, got, values[i])
		}
	}
}

func TestValueEditor_MigrateWithValue_MultiChunk_PerChunkValues(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	n := 600 // spans multiple chunks under ent.DefaultConfig's chunk capacity
	ids := spawnAll(m, accessSpec, n)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	values := make([]mVelocity, n)
	for i := range values {
		values[i] = mVelocity{VX: float64(i), VY: float64(i)}
	}
	applyByChunksWithValue(m, vm, ids, values)

	for i, id := range ids {
		if got := readVel(m, velDef, id); got != values[i] {
			t.Errorf("entity ids[%d]: velocity = %+v; want %+v", i, got, values[i])
		}
	}
}

func TestValueEditor_MigrateWithValue_ComponentAlreadyPresent_WritesInPlace(t *testing.T) {
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
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	values := []mVelocity{{VX: 5, VY: 5}, {VX: 6, VY: 6}}
	applyByChunksWithValue(m, vm, ids, values)

	for i, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after MigrateWithValue", id)
		}
		if entry.ArchID != srcArchID {
			t.Errorf("entity %v: archetype changed even though composition was already the target", id)
		}
		if got := readVel(m, velDef, id); got != values[i] {
			t.Errorf("entity %v: velocity = %+v; want %+v", id, got, values[i])
		}
	}
}

func TestValueEditor_MigrateWithValue_AddDelSameComponent_NoOpWithoutCrash(t *testing.T) {
	// Add(vel) + Del(vel) on entities that never had vel cancels out to the
	// unchanged source composition in resolveDst, but the vel column never
	// existed in this table — MigrateWithValue must not dereference a nil
	// column pointer.
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 2)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])), comp.Del[mVelocity]())
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	values := []mVelocity{{VX: 1, VY: 1}, {VX: 2, VY: 2}}
	applyByChunksWithValue(m, vm, ids, values) // must not panic

	for _, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after MigrateWithValue", id)
		}
		if entry.ArchID != srcArchID {
			t.Errorf("entity %v: archetype changed; composition should be net-unchanged", id)
		}
	}
}

func TestValueEditor_ValueType(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	if got, want := vm.ValueType(), reflect.TypeFor[mVelocity](); got != want {
		t.Errorf("ValueType() = %v; want %v", got, want)
	}
}

func TestValueEditor_MigrateWithValue_StaleVersion_AllEntitiesDead_NoOp(t *testing.T) {
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

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

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

	values := []mVelocity{{VX: 1, VY: 1}, {VX: 2, VY: 2}, {VX: 3, VY: 3}}
	vm.MigrateWithValue(snap, ids, unsafe.Pointer(unsafe.SliceData(values))) // must not panic; nothing left to write

	if got := srcTable.Len(); got != 0 {
		t.Errorf("src Table.Len = %d; expected 0", got)
	}
}

func TestValueEditor_MigrateWithValue_ZeroSizeComponentAlreadyPresent_NoOp(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)
	tagDef := mi.Intern(reflect.TypeFor[mTag]())

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	_ = accessSpec.Tag(tagDef.ID)
	ids := spawnAll(m, accessSpec, 2)

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mTag])))
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	applyByChunksWithValue(m, vm, ids, []mTag{{}, {}}) // must not panic; zero-size, nothing to write

	for _, id := range ids {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after MigrateWithValue", id)
		}
		if entry.ArchID != srcArchID {
			t.Errorf("entity %v: archetype changed even though composition was already the target", id)
		}
	}
}

func TestValueEditor_MigrateWithValue_PartialBatch_SurvivorsRelocated(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, velDef := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 4) // ids[0..3] seeded at slots 0..3

	entry0, _ := m.AddressBook.Get(ids[0])
	srcArchID := entry0.ArchID

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])))
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	// Only ids[0] and ids[2] migrate — ids[1], ids[3] stay behind and must be
	// relocated by the source-compaction pass (the srcMoves address-book loop).
	values := []mVelocity{{VX: 1, VY: 1}, {VX: 3, VY: 3}}
	applyByChunksWithValue(m, vm, []uid.UID64{ids[0], ids[2]}, values)

	if got := m.ArchCatalog.Archetypes[srcArchID].Table.Len(); got != 2 {
		t.Errorf("src Table.Len = %d; expected 2", got)
	}

	want := map[uid.UID64]mVelocity{ids[0]: values[0], ids[2]: values[1]}
	for id, wantVel := range want {
		entry, ok := m.AddressBook.Get(id)
		if !ok {
			t.Fatalf("entity %v missing after MigrateWithValue", id)
		}
		if entry.ArchID == srcArchID {
			t.Errorf("entity %v: still in source archetype", id)
		}
		if got := readVel(m, velDef, id); got != wantVel {
			t.Errorf("entity %v: velocity = %+v; want %+v", id, got, wantVel)
		}
	}

	for _, i := range []int{1, 3} {
		entry, ok := m.AddressBook.Get(ids[i])
		if !ok {
			t.Fatalf("ids[%d] missing after MigrateWithValue (should have survived)", i)
		}
		if entry.ArchID != srcArchID {
			t.Errorf("ids[%d]: expected to remain in source archetype", i)
		}
	}
}

func TestValueEditor_MigrateWithValue_Unlink_RemovesEntities(t *testing.T) {
	m := newMgr()
	var mi comp.DefIndex
	mi.Init()
	posDef, _ := internDefs(&mi)

	var accessSpec comp.AccessSpec
	_ = accessSpec.Comp(posDef)
	ids := spawnAll(m, accessSpec, 2)

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(new(iter.ArrayRef[mVelocity])), comp.Del[mVelocity](), comp.Del[mPosition]())
	vm := ent.NewValueEditor(&m.AddressBook, &m.ArchCatalog, editSpec)

	values := []mVelocity{{VX: 1, VY: 1}, {VX: 2, VY: 2}}
	applyByChunksWithValue(m, vm, ids, values)

	for _, id := range ids {
		if _, ok := m.AddressBook.Get(id); ok {
			t.Errorf("entity %v should be removed (unlinked) but still exists", id)
		}
	}
}
