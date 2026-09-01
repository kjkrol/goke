package reg_test

import (
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/internal/query"
	"github.com/kjkrol/goke/v3/internal/reg"
	"github.com/kjkrol/goke/v3/iter"
)

func TestDefaultConfig(t *testing.T) {
	cfg := reg.DefaultConfig()

	if cfg.Entity != ent.DefaultConfig() {
		t.Errorf("expected Entity sub-config to match ent.DefaultConfig(), got %+v", cfg.Entity)
	}
	if cfg.Matcher != query.DefaultConfig() {
		t.Errorf("expected Matcher sub-config to match query.DefaultConfig(), got %+v", cfg.Matcher)
	}
}

func TestRegistry_RegComp(t *testing.T) {
	r := newRegistry(t)

	id1 := r.RegComp(reflect.TypeFor[Position]())
	id1Again := r.RegComp(reflect.TypeFor[Position]())
	id2 := r.RegComp(reflect.TypeFor[Velocity]())

	if id1 != id1Again {
		t.Errorf("expected RegComp to be idempotent for the same type, got %d then %d", id1, id1Again)
	}
	if id1 == id2 {
		t.Error("expected distinct types to get distinct IDs")
	}
}

func TestRegistry_CreateFactory(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(2)

	total := 0
	for factory.Next() {
		positions := pos.Slice(&factory.Cursor)
		for i := range positions {
			positions[i] = Position{X: float64(total + i)}
		}
		total += len(factory.IDs)
	}
	if total != 2 {
		t.Errorf("expected 2 entities created, got %d", total)
	}
}

func TestRegistry_CreateFactory_PanicsOnDelOpt(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	defer func() {
		if recover() == nil {
			t.Error("expected CreateFactory to panic when given a Del opt")
		}
	}()
	r.CreateFactory(comp.Del[Position]())
}

func TestRegistry_CreateFactory_PanicsOnDuplicateTagAddOpt(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Tag]())

	defer func() {
		if recover() == nil {
			t.Error("expected CreateFactory to panic when the same tag is added twice")
		}
	}()
	r.CreateFactory(comp.Add(new(iter.ArrayRef[Tag])), comp.Add(new(iter.ArrayRef[Tag])))
}

func TestRegistry_CreateFactory_PanicsOnDuplicateCompAddOpt(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	defer func() {
		if recover() == nil {
			t.Error("expected CreateFactory to panic when the same component is added twice")
		}
	}()
	var p1, p2 iter.ArrayRef[Position]
	r.CreateFactory(comp.Add(&p1), comp.Add(&p2))
}

func TestRegistry_CreateFactory_AcceptsZeroSizeAddOpt(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())
	r.RegComp(reflect.TypeFor[Tag]())
	r.RegComp(reflect.TypeFor[Velocity]())

	var pos iter.ArrayRef[Position]
	var vel iter.ArrayRef[Velocity]
	factory := r.CreateFactory(comp.Add(&pos), comp.Add(new(iter.ArrayRef[Tag])), comp.Add(&vel))
	factory.Create(1)
	factory.Next()
	pos.Slice(&factory.Cursor)[0] = Position{X: 1, Y: 2}
	vel.Slice(&factory.Cursor)[0] = Velocity{VX: 3, VY: 4}
	id := factory.IDs[0]

	var trackedPos iter.ArrayRef[Position]
	var trackedVel iter.ArrayRef[Velocity]
	matcher := r.AddMatcher(comp.Include[Tag](), comp.Track(&trackedPos), comp.Track(&trackedVel))
	matcher.All()
	found := false
	for matcher.Next() {
		positions := trackedPos.Slice(&matcher.Cursor)
		velocities := trackedVel.Slice(&matcher.Cursor)
		for i, gotID := range matcher.Cursor.IDs {
			if gotID != id {
				continue
			}
			found = true
			if positions[i] != (Position{X: 1, Y: 2}) {
				t.Errorf("Position = %+v, want {1 2}", positions[i])
			}
			if velocities[i] != (Velocity{VX: 3, VY: 4}) {
				t.Errorf("Velocity = %+v, want {3 4}", velocities[i])
			}
		}
	}
	if !found {
		t.Error("expected the spawned entity to match a query requiring the zero-size tag component")
	}
}

func TestRegistry_CreateFactory_SliceOnTagBoundColPanics(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Tag]())

	var tag iter.ArrayRef[Tag]
	factory := r.CreateFactory(comp.Add(&tag))
	factory.Create(1)
	factory.Next()

	defer func() {
		if recover() == nil {
			t.Error("expected Slice on a tag-bound col to panic")
		}
	}()
	tag.Slice(&factory.Cursor)
}

func TestRegistry_Remove(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	if !r.Remove(id) {
		t.Error("expected Remove to succeed for a known entity")
	}
	if r.Remove(id) {
		t.Error("expected Remove to return false for an already-removed entity")
	}
	if r.Remove(uid.UID64(999)) {
		t.Error("expected Remove to return false for an unknown entity")
	}
}

func TestRegistry_UpsertAndRemoveComp(t *testing.T) {
	r := newRegistry(t)
	posID := r.RegComp(reflect.TypeFor[Position]())
	velID := r.RegComp(reflect.TypeFor[Velocity]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	ptr, err := r.UpsertComp(id, velID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ptr == nil {
		t.Fatal("expected a non-nil pointer for a data component")
	}
	*(*Velocity)(ptr) = Velocity{VX: 1, VY: 2}

	if err := r.RemoveComp(id, posID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := r.UpsertComp(uid.UID64(999), velID); err == nil {
		t.Error("expected an error for UpsertComp on an unknown entity")
	}
	if err := r.RemoveComp(uid.UID64(999), velID); err == nil {
		t.Error("expected an error for RemoveComp on an unknown entity")
	}
}

func TestRegistry_CreateEditor(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())
	r.RegComp(reflect.TypeFor[Velocity]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(2)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	entry0, ok := r.EntityManager.AddressBook.Get(ids[0])
	if !ok {
		t.Fatal("expected entity to be present in AddressBook")
	}
	srcTable := &r.EntityManager.ArchCatalog.Archetypes[entry0.ArchID].Table

	var vel iter.ArrayRef[Velocity]
	editor := r.CreateEditor(comp.Add(&vel))

	snap := bulk.ChunkSnapshot{
		ArchID:      entry0.ArchID,
		ChunkPtr:    entry0.ChunkPtr,
		ChunkIdx:    srcTable.ChunkIdxByPtr(entry0.ChunkPtr),
		TableVer:    srcTable.Version(),
		SlotAligned: true,
	}
	editor.Migrate(snap, ids)

	entryAfter, ok := r.EntityManager.AddressBook.Get(ids[0])
	if !ok {
		t.Fatal("entity missing after migration")
	}
	if entryAfter.ArchID == entry0.ArchID {
		t.Error("expected entity to move to a new archetype after Migrate")
	}
}

func TestRegistry_Remover(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(2)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	entry0, ok := r.EntityManager.AddressBook.Get(ids[0])
	if !ok {
		t.Fatal("expected entity to be present in AddressBook")
	}
	srcTable := &r.EntityManager.ArchCatalog.Archetypes[entry0.ArchID].Table

	remover := r.Remover()

	snap := bulk.ChunkSnapshot{
		ArchID:      entry0.ArchID,
		ChunkPtr:    entry0.ChunkPtr,
		ChunkIdx:    srcTable.ChunkIdxByPtr(entry0.ChunkPtr),
		TableVer:    srcTable.Version(),
		SlotAligned: true,
	}
	remover.Migrate(snap, ids)

	for _, id := range ids {
		if _, ok := r.EntityManager.AddressBook.Get(id); ok {
			t.Errorf("entity %v: expected removed by Remover, still present", id)
		}
	}
}

func TestRegistry_CreateValueEditor(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())
	r.RegComp(reflect.TypeFor[Velocity]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(2)
	var ids []uid.UID64
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}

	entry0, ok := r.EntityManager.AddressBook.Get(ids[0])
	if !ok {
		t.Fatal("expected entity to be present in AddressBook")
	}
	srcTable := &r.EntityManager.ArchCatalog.Archetypes[entry0.ArchID].Table

	var vel iter.ArrayRef[Velocity]
	vm := r.CreateValueEditor(comp.Add(&vel))

	if got := vm.ValueType(); got != reflect.TypeFor[Velocity]() {
		t.Errorf("ValueType() = %v; want %v", got, reflect.TypeFor[Velocity]())
	}

	snap := bulk.ChunkSnapshot{
		ArchID:      entry0.ArchID,
		ChunkPtr:    entry0.ChunkPtr,
		ChunkIdx:    srcTable.ChunkIdxByPtr(entry0.ChunkPtr),
		TableVer:    srcTable.Version(),
		SlotAligned: true,
	}
	values := []Velocity{{VX: 1, VY: 1}, {VX: 2, VY: 2}}
	vm.MigrateWithValue(snap, ids, unsafe.Pointer(unsafe.SliceData(values)))

	entryAfter, ok := r.EntityManager.AddressBook.Get(ids[0])
	if !ok {
		t.Fatal("entity missing after MigrateWithValue")
	}
	if entryAfter.ArchID == entry0.ArchID {
		t.Error("expected entity to move to a new archetype after MigrateWithValue")
	}
}

func TestRegistry_CreateValueEditor_PanicsOnNotExactlyOneAdd(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when no component is added")
		}
	}()
	r.CreateValueEditor()
}

func TestRegistry_Reset(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	r.Reset()

	if r.Remove(id) {
		t.Error("expected the entity to no longer exist after Reset")
	}

	// Registry must remain usable after Reset — including re-registering the
	// same Go type (CompDefIndex.Reset must clear the type->ID map too).
	newID := r.RegComp(reflect.TypeFor[Position]())
	if newID != 0 {
		t.Errorf("expected first registration after Reset to get ID 0, got %d", newID)
	}
}

func TestRegistry_Save_PanicsWithoutPause(t *testing.T) {
	r := newRegistry(t)

	defer func() {
		if recover() == nil {
			t.Error("expected Save to panic without a prior Pause()")
		}
	}()
	_ = r.Save(filepath.Join(t.TempDir(), "save.bin"))
}

func TestRegistry_Save_ReturnsFileError(t *testing.T) {
	r := newRegistry(t)
	r.Pause()

	err := r.Save(filepath.Join(t.TempDir(), "no-such-dir", "save.bin"))
	if err == nil {
		t.Error("expected an error when the target directory doesn't exist")
	}
}

func TestRegistry_Load_ReturnsFileError(t *testing.T) {
	r := newRegistry(t)

	err := r.Load(filepath.Join(t.TempDir(), "does-not-exist.bin"), nil)
	if err == nil {
		t.Error("expected an error when the save file doesn't exist")
	}
}

func TestRegistry_Load_PanicsIfNotFresh(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	defer func() {
		if recover() == nil {
			t.Error("expected Load to panic when called after other registration")
		}
	}()
	_ = r.Load(filepath.Join(t.TempDir(), "save.bin"), []reg.CompToken{reg.LoadComp[Position]()})
}

func TestRegistry_SaveLoad_RoundTrip(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())

	var pos iter.ArrayRef[Position]
	factory := r.CreateFactory(comp.Add(&pos))
	factory.Create(1)
	factory.Next()
	pos.Slice(&factory.Cursor)[0] = Position{X: 1, Y: 2}

	path := filepath.Join(t.TempDir(), "save.bin")

	r.Pause()
	if !r.Paused() {
		t.Fatal("expected Paused() true after Pause")
	}
	if err := r.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	r2 := newRegistry(t)
	if err := r2.Load(path, []reg.CompToken{reg.LoadComp[Position]()}); err != nil {
		t.Fatalf("Load: %v", err)
	}

	var trackedPos iter.ArrayRef[Position]
	matcher := r2.AddMatcher(comp.Track(&trackedPos))
	matcher.All()
	found := false
	for matcher.Next() {
		positions := trackedPos.Slice(&matcher.Cursor)
		for i := range positions {
			found = true
			if positions[i] != (Position{X: 1, Y: 2}) {
				t.Errorf("Position = %+v, want {1 2}", positions[i])
			}
		}
	}
	if !found {
		t.Fatal("expected the loaded entity to be found")
	}
}

func TestRegistry_PauseResume_Idempotent(t *testing.T) {
	r := newRegistry(t)

	r.Resume() // no-op when not paused
	if r.Paused() {
		t.Fatal("expected Paused() false initially")
	}

	r.Pause()
	r.Pause() // idempotent
	if !r.Paused() {
		t.Fatal("expected Paused() true after Pause")
	}

	r.Resume()
	r.Resume() // idempotent
	if r.Paused() {
		t.Fatal("expected Paused() false after Resume")
	}
}
