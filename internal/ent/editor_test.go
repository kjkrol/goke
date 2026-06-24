package ent_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

func TestEditor_Update_InvalidEntity(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	var col iter.ArrayRef[Position]
	var spec comp.EditSpec
	spec.Init(&mi, comp.Add(&col))
	editor := m.CreateEditor(spec)

	if editor.Update(uid.UID64(999)) {
		t.Error("expected Update to return false for an unknown entity")
	}
}

func TestEditor_Update_AddComponent(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	tagDef := mi.Intern(reflect.TypeFor[Tag]())
	posDef := mi.Intern(reflect.TypeFor[Position]())

	factory := m.CreateFactory(tagAccessSpec(tagDef.ID))
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	var posCol iter.ArrayRef[Position]
	var spec comp.EditSpec
	spec.Init(&mi, comp.Add(&posCol))
	editor := m.CreateEditor(spec)

	if !editor.Update(id) {
		t.Fatal("expected Update to succeed")
	}
	posCol.At(&editor.Cursor).X = 5
	posCol.At(&editor.Cursor).Y = 6

	entry, _ := m.AddressBook.Get(id)
	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(tagDef.ID) {
		t.Error("expected the original Tag bit to survive the add")
	}
	got := *(*Position)(m.ArchCatalog.Archetypes[entry.ArchId].Table.ComponentAt(entry.Pos, posDef.ID))
	if got != (Position{X: 5, Y: 6}) {
		t.Errorf("expected written Position{5,6}, got %+v", got)
	}
}

func TestEditor_Update_RemoveComponent(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())
	velDef := mi.Intern(reflect.TypeFor[Velocity]())

	spec := comp.AccessSpec{}
	_ = spec.Comp(posDef)
	_ = spec.Comp(velDef)
	factory := m.CreateFactory(spec)
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Del[Velocity]())
	editor := m.CreateEditor(editSpec)

	if !editor.Update(id) {
		t.Fatal("expected Update to succeed")
	}

	entry, ok := m.AddressBook.Get(id)
	if !ok {
		t.Fatal("expected entity to survive (Position remains)")
	}
	if m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(velDef.ID) {
		t.Error("expected Velocity bit to be cleared")
	}
	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(posDef.ID) {
		t.Error("expected Position bit to survive")
	}
}

func TestEditor_Update_AddAndRemoveTogether(t *testing.T) {
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

	var velCol iter.ArrayRef[Velocity]
	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Add(&velCol), comp.Del[Position]())
	editor := m.CreateEditor(editSpec)

	if !editor.Update(id) {
		t.Fatal("expected Update to succeed")
	}
	velCol.At(&editor.Cursor).VX = 1

	entry, ok := m.AddressBook.Get(id)
	if !ok {
		t.Fatal("expected entity to survive (Velocity remains)")
	}
	if m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(posDef.ID) {
		t.Error("expected Position bit to be cleared")
	}
	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(velDef.ID) {
		t.Error("expected Velocity bit to be set")
	}
}

func TestEditor_Update_UnlinksWhenResultIsEmpty(t *testing.T) {
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

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Del[Position]())
	editor := m.CreateEditor(editSpec)

	if !editor.Update(id) {
		t.Fatal("expected Update to succeed (the unlink path still returns true)")
	}

	if _, ok := m.AddressBook.Get(id); ok {
		t.Error("expected the entity to be fully unlinked")
	}
}

// A pure-removal Editor must still work correctly even though it never
// positions a cursor for writes.
func TestEditor_Update_DelOnlyEditorWorks(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())
	velDef := mi.Intern(reflect.TypeFor[Velocity]())

	spec := comp.AccessSpec{}
	_ = spec.Comp(posDef)
	_ = spec.Comp(velDef)
	factory := m.CreateFactory(spec)
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]

	var editSpec comp.EditSpec
	editSpec.Init(&mi, comp.Del[Velocity]())
	editor := m.CreateEditor(editSpec)

	for i := 0; i < 3; i++ {
		if !editor.Update(id) && i == 0 {
			t.Fatal("expected first Update to succeed")
		}
	}

	entry, ok := m.AddressBook.Get(id)
	if !ok {
		t.Fatal("expected entity to survive")
	}
	if m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(velDef.ID) {
		t.Error("expected Velocity to stay removed")
	}
}

// Reusing the same Editor across entities that end up in different target
// archetypes must re-bake offsets correctly each time, never reusing a
// stale cache from a different archetype.
func TestEditor_Update_AcrossDifferentTargetArchetypes(t *testing.T) {
	m := newManager()
	var mi comp.DefIndex
	mi.Init()
	posDef := mi.Intern(reflect.TypeFor[Position]())
	velDef := mi.Intern(reflect.TypeFor[Velocity]())

	// Entity A: starts with just Position.
	specA := comp.AccessSpec{}
	_ = specA.Comp(posDef)
	fa := m.CreateFactory(specA)
	fa.Create(1)
	fa.Next()
	idA := fa.IDs[0]

	// Entity B: starts with Position + Velocity (a different archetype).
	specB := comp.AccessSpec{}
	_ = specB.Comp(posDef)
	_ = specB.Comp(velDef)
	fb := m.CreateFactory(specB)
	fb.Create(1)
	fb.Next()
	idB := fb.IDs[0]

	// Both get a Velocity added — A migrates into the {Pos,Vel} archetype
	// (idempotent no-op for B, which already has it).
	var velCol iter.ArrayRef[Velocity]
	var spec comp.EditSpec
	spec.Init(&mi, comp.Add(&velCol))
	editor := m.CreateEditor(spec)

	for i := 0; i < 3; i++ {
		if !editor.Update(idA) {
			t.Fatalf("iter %d: expected Update(idA) to succeed", i)
		}
		velCol.At(&editor.Cursor).VX = 100

		if !editor.Update(idB) {
			t.Fatalf("iter %d: expected Update(idB) to succeed", i)
		}
		velCol.At(&editor.Cursor).VX = 200

		entryA, _ := m.AddressBook.Get(idA)
		gotA := *(*Velocity)(m.ArchCatalog.Archetypes[entryA.ArchId].Table.ComponentAt(entryA.Pos, velDef.ID))
		if gotA.VX != 100 {
			t.Errorf("iter %d: expected idA.VX == 100, got %v", i, gotA.VX)
		}

		entryB, _ := m.AddressBook.Get(idB)
		gotB := *(*Velocity)(m.ArchCatalog.Archetypes[entryB.ArchId].Table.ComponentAt(entryB.Pos, velDef.ID))
		if gotB.VX != 200 {
			t.Errorf("iter %d: expected idB.VX == 200, got %v", i, gotB.VX)
		}
	}
}
