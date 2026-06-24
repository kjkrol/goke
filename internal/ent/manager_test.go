package ent_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/ent"
)

// Init must invoke the caller's onArchetypeCreated callback for every new
// archetype, in addition to wiring up the IDSeeder.
func TestManager_Init_InvokesOnArchetypeCreatedCallback(t *testing.T) {
	var created []arch.ID
	var m ent.Manager
	m.Init(ent.DefaultConfig(), func(a *arch.Archetype) {
		created = append(created, a.Id)
	})

	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(1)
	factory.Next()

	if len(created) != 1 {
		t.Fatalf("expected the callback to fire once for the new tag archetype, got %d calls", len(created))
	}
}

func TestManager_RemoveUnknownEntity(t *testing.T) {
	m := newManager()

	if m.Remove(uid.UID64(999)) {
		t.Error("expected Remove to return false for an unknown entity")
	}
}

func TestManager_CreateFactoryAndRemove(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	if _, ok := m.AddressBook.Get(id); !ok {
		t.Fatal("expected entity to be addressable after creation")
	}

	if !m.Remove(id) {
		t.Error("expected Remove to succeed for a known entity")
	}
	if _, ok := m.AddressBook.Get(id); ok {
		t.Error("expected entity to be gone after Remove")
	}
}

func TestManager_UpsertComp_InvalidEntity(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())

	if _, err := m.UpsertComp(uid.UID64(999), posDef); err == nil {
		t.Error("expected an error for an unknown entity")
	}
}

func TestManager_UpsertComp_NewComponent(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())
	posDef := mi.Intern(reflect.TypeFor[Position]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	ptr, err := m.UpsertComp(id, posDef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ptr == nil {
		t.Fatal("expected a non-nil pointer for a data component")
	}
	*(*Position)(ptr) = Position{X: 1, Y: 2}

	entry, _ := m.AddressBook.Get(id)
	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(posDef.ID) {
		t.Error("expected the entity's archetype to have the Position bit set")
	}
}

func TestManager_UpsertComp_TagComponent(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())
	tagDef := mi.Intern(reflect.TypeFor[Tag]())

	spec := comp.AccessSpec{}
	_ = spec.Comp(posDef)
	factory := m.CreateFactory(spec)
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	ptr, err := m.UpsertComp(id, tagDef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ptr != nil {
		t.Errorf("expected a nil pointer for a zero-size tag, got %v", ptr)
	}

	entry, _ := m.AddressBook.Get(id)
	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(tagDef.ID) {
		t.Error("expected the entity's archetype to have the Tag bit set")
	}
}

func TestManager_UpsertComp_Idempotent(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())
	posDef := mi.Intern(reflect.TypeFor[Position]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	ptr1, _ := m.UpsertComp(id, posDef)
	*(*Position)(ptr1) = Position{X: 7, Y: 8}
	entryBefore, _ := m.AddressBook.Get(id)

	ptr2, err := m.UpsertComp(id, posDef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entryAfter, _ := m.AddressBook.Get(id)
	if entryBefore.ArchId != entryAfter.ArchId || entryBefore.Pos != entryAfter.Pos {
		t.Error("re-upserting an already-present component should not move the entity")
	}
	if got := *(*Position)(ptr2); got != (Position{X: 7, Y: 8}) {
		t.Errorf("expected data to survive idempotent upsert, got %+v", got)
	}
}

func TestManager_RemoveComp_InvalidEntity(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())

	if err := m.RemoveComp(uid.UID64(999), posDef); err == nil {
		t.Error("expected an error for an unknown entity")
	}
}

func TestManager_RemoveComp_NotPresentIsNoOp(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())
	posDef := mi.Intern(reflect.TypeFor[Position]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]
	entryBefore, _ := m.AddressBook.Get(id)

	if err := m.RemoveComp(id, posDef); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entryAfter, _ := m.AddressBook.Get(id)
	if entryBefore.ArchId != entryAfter.ArchId {
		t.Error("removing an absent component should not move the entity")
	}
}

func TestManager_RemoveComp_MigratesWithoutUnlinking(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())
	velDef := mi.Intern(reflect.TypeFor[Velocity]())

	spec := comp.AccessSpec{}
	_ = spec.Comp(posDef)
	factory := m.CreateFactory(spec)
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	posPtr, _ := m.UpsertComp(id, posDef)
	*(*Position)(posPtr) = Position{X: 3, Y: 4}
	_, _ = m.UpsertComp(id, velDef)

	if err := m.RemoveComp(id, velDef); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, ok := m.AddressBook.Get(id)
	if !ok {
		t.Fatal("expected entity to still exist after removing one of two components")
	}
	if m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(velDef.ID) {
		t.Error("expected Velocity bit to be cleared")
	}
	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(posDef.ID) {
		t.Error("expected Position bit to survive the migration")
	}

	got := *(*Position)(m.ArchCatalog.Archetypes[entry.ArchId].Table.ComponentAt(entry.Pos, posDef.ID))
	if got != (Position{X: 3, Y: 4}) {
		t.Errorf("expected Position data to survive migration, got %+v", got)
	}
}

func TestManager_RemoveComp_UnlinksWhenLastComponentRemoved(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())

	spec := comp.AccessSpec{}
	_ = spec.Comp(posDef)
	factory := m.CreateFactory(spec)
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	if err := m.RemoveComp(id, posDef); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := m.AddressBook.Get(id); ok {
		t.Error("expected the entity to be fully unlinked after removing its only component")
	}
}

// Removing a non-last entity from an archetype's table must swap the last
// entity into the vacated slot, and Manager must update that swapped
// entity's address accordingly.
func TestManager_Remove_SwapPopUpdatesDisplacedEntity(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())

	spec := comp.AccessSpec{}
	_ = spec.Comp(posDef)
	factory := m.CreateFactory(spec)
	factory.Create(3)
	factory.Next()
	ids := append([]uid.UID64{}, factory.IDs...)
	if len(ids) != 3 {
		t.Fatalf("expected 3 entities, got %d", len(ids))
	}

	entryLastBefore, _ := m.AddressBook.Get(ids[2])
	if entryLastBefore.Pos.Slot != 2 {
		t.Fatalf("setup error: expected last entity at slot 2, got %d", entryLastBefore.Pos.Slot)
	}

	// Remove the first entity — the last one must swap into its slot.
	if !m.Remove(ids[0]) {
		t.Fatal("expected Remove to succeed")
	}

	entryLastAfter, ok := m.AddressBook.Get(ids[2])
	if !ok {
		t.Fatal("expected the displaced entity to remain addressable")
	}
	if entryLastAfter.Pos.Slot != 0 {
		t.Errorf("expected swapped entity to move to slot 0, got %d", entryLastAfter.Pos.Slot)
	}
}

// migrateEntity (exercised via UpsertComp) must also keep a swapped entity's
// address correct when the migrating entity isn't the last in its source
// archetype's table.
func TestManager_UpsertComp_SwapPopUpdatesDisplacedEntity(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())
	velDef := mi.Intern(reflect.TypeFor[Velocity]())

	spec := comp.AccessSpec{}
	_ = spec.Comp(posDef)
	factory := m.CreateFactory(spec)
	factory.Create(3)
	factory.Next()
	ids := append([]uid.UID64{}, factory.IDs...)

	// Migrate the first entity away by adding Velocity — entity at the last
	// slot of the source archetype must swap into its old slot.
	if _, err := m.UpsertComp(ids[0], velDef); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entryLast, ok := m.AddressBook.Get(ids[2])
	if !ok {
		t.Fatal("expected the displaced entity to remain addressable")
	}
	if entryLast.Pos.Slot != 0 {
		t.Errorf("expected swapped entity to move to slot 0, got %d", entryLast.Pos.Slot)
	}
	if entryLast.ArchId == arch.NullID {
		t.Error("expected the displaced entity to still have a valid archetype")
	}
}

func TestManager_Reset(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	m.Reset()

	if _, ok := m.AddressBook.Get(id); ok {
		t.Error("expected no entries to survive Reset")
	}

	// Manager must remain usable after Reset.
	factory2 := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory2.Create(1)
	factory2.Next()
	if len(factory2.IDs) != 1 {
		t.Error("expected Manager to be usable again after Reset")
	}
}
