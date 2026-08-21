package persist_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/internal/persist"
	"github.com/kjkrol/goke/v3/iter"
)

type Position struct{ X, Y float32 }
type Velocity struct{ X, Y float32 }
type Tag struct{}
type Name struct{ Value string }

type stamp struct{ V int64 }

func (s stamp) MarshalBinary() ([]byte, error) { return fmt.Appendf(nil, "%d", s.V), nil }
func (s *stamp) UnmarshalBinary(data []byte) error {
	_, err := fmt.Sscanf(string(data), "%d", &s.V)
	return err
}

func newTestManager() *ent.Manager {
	var m ent.Manager
	m.Init(ent.DefaultConfig(), nil)
	return &m
}

// req builds a persist.CompRequest that registers T into di on request,
// validating the recorded size when the save file has an entry for it.
func req[T any](di *comp.DefIndex) persist.CompRequest {
	t := reflect.TypeFor[T]()
	return persist.CompRequest{
		Name: t.String(),
		Register: func(wantSize *uint32) error {
			if wantSize != nil && uint32(t.Size()) != *wantSize {
				return fmt.Errorf("size mismatch for %s: file has %d, current type has %d", t, *wantSize, t.Size())
			}
			di.Intern(t)
			return nil
		},
	}
}

func componentAt[T any](m *ent.Manager, id uid.UID64, compID comp.ID) (T, bool) {
	entry, ok := m.AddressBook.Get(id)
	if !ok {
		var zero T
		return zero, false
	}
	ptr := m.ArchCatalog.Archetypes[entry.ArchID].Table.ComponentAt(entry.ChunkPtr, entry.Slot, compID)
	return *(*T)(ptr), true
}

func TestSaveLoad_RoundTrip_EmptyWorld(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()

	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()
	if err := persist.Load(&buf, &di2, &m2.AddressBook, &m2.ArchCatalog, nil); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestSaveLoad_RoundTrip_MultipleArchetypes(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()

	var posCol iter.ArrayRef[Position]
	var velCol iter.ArrayRef[Velocity]
	var nameCol iter.ArrayRef[Name]
	var stampCol iter.ArrayRef[stamp]

	posID := di.Intern(reflect.TypeFor[Position]()).ID
	tagID := di.Intern(reflect.TypeFor[Tag]()).ID
	nameID := di.Intern(reflect.TypeFor[Name]()).ID
	stampID := di.Intern(reflect.TypeFor[stamp]()).ID
	_ = di.Intern(reflect.TypeFor[Velocity]()).ID

	// Archetype A: Position + Velocity (both POD).
	var specA comp.AccessSpec
	specA.Init(&di, comp.Track(&posCol), comp.Track(&velCol))
	factoryA := m.CreateFactory(specA)
	factoryA.Create(3)
	var idsA []uid.UID64
	for factoryA.Next() {
		positions := posCol.Slice(&factoryA.Cursor)
		velocities := velCol.Slice(&factoryA.Cursor)
		for i := range positions {
			positions[i] = Position{X: float32(i), Y: float32(i) * 2}
			velocities[i] = Velocity{X: 1, Y: -1}
		}
		idsA = append(idsA, factoryA.IDs...)
	}

	// Archetype B: Position + Tag (tag has zero size — mask-only).
	var specB comp.AccessSpec
	specB.Init(&di, comp.Track(&posCol))
	if err := specB.Tag(tagID); err != nil {
		t.Fatal(err)
	}
	factoryB := m.CreateFactory(specB)
	factoryB.Create(2)
	var idsB []uid.UID64
	for factoryB.Next() {
		positions := posCol.Slice(&factoryB.Cursor)
		for i := range positions {
			positions[i] = Position{X: 100 + float32(i), Y: 0}
		}
		idsB = append(idsB, factoryB.IDs...)
	}

	// Archetype C: Name (string) + stamp (BinaryMarshaler).
	var specC comp.AccessSpec
	specC.Init(&di, comp.Track(&nameCol), comp.Track(&stampCol))
	factoryC := m.CreateFactory(specC)
	factoryC.Create(2)
	var idsC []uid.UID64
	for factoryC.Next() {
		names := nameCol.Slice(&factoryC.Cursor)
		stamps := stampCol.Slice(&factoryC.Cursor)
		for i := range names {
			names[i] = Name{Value: fmt.Sprintf("entity-%d", i)}
			stamps[i] = stamp{V: int64(i) * 1000}
		}
		idsC = append(idsC, factoryC.IDs...)
	}

	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()
	comps := []persist.CompRequest{
		req[Position](&di2), req[Velocity](&di2), req[Tag](&di2), req[Name](&di2), req[stamp](&di2),
	}
	if err := persist.Load(&buf, &di2, &m2.AddressBook, &m2.ArchCatalog, comps); err != nil {
		t.Fatalf("Load: %v", err)
	}

	posID2, _ := di2.ByType(reflect.TypeFor[Position]())
	velID2, _ := di2.ByType(reflect.TypeFor[Velocity]())
	nameID2, _ := di2.ByType(reflect.TypeFor[Name]())
	stampID2, _ := di2.ByType(reflect.TypeFor[stamp]())

	// comp.ID assignment order must be reproduced exactly (Position,Tag,Name,stamp,Velocity).
	if posID2.ID != posID || nameID2.ID != nameID || stampID2.ID != stampID {
		t.Fatalf("comp.ID assignment diverged: pos %d/%d name %d/%d stamp %d/%d",
			posID2.ID, posID, nameID2.ID, nameID, stampID2.ID, stampID)
	}

	for i, id := range idsA {
		pos, ok := componentAt[Position](m2, id, posID2.ID)
		if !ok {
			t.Fatalf("archA entity %d (%v) missing after Load", i, id)
		}
		if want := (Position{X: float32(i), Y: float32(i) * 2}); pos != want {
			t.Errorf("archA entity %d: Position = %+v, want %+v", i, pos, want)
		}
		vel, ok := componentAt[Velocity](m2, id, velID2.ID)
		if !ok || vel != (Velocity{X: 1, Y: -1}) {
			t.Errorf("archA entity %d: Velocity = %+v, ok=%v", i, vel, ok)
		}
	}

	for i, id := range idsB {
		pos, ok := componentAt[Position](m2, id, posID2.ID)
		if !ok {
			t.Fatalf("archB entity %d (%v) missing after Load", i, id)
		}
		if want := (Position{X: 100 + float32(i), Y: 0}); pos != want {
			t.Errorf("archB entity %d: Position = %+v, want %+v", i, pos, want)
		}
	}

	for i, id := range idsC {
		name, ok := componentAt[Name](m2, id, nameID2.ID)
		if !ok || name.Value != fmt.Sprintf("entity-%d", i) {
			t.Errorf("archC entity %d: Name = %+v, ok=%v", i, name, ok)
		}
		st, ok := componentAt[stamp](m2, id, stampID2.ID)
		if !ok || st.V != int64(i)*1000 {
			t.Errorf("archC entity %d: stamp = %+v, ok=%v", i, st, ok)
		}
	}
}

func TestSaveLoad_LoadCompOrderIsIndependent(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()

	var posCol iter.ArrayRef[Position]
	var velCol iter.ArrayRef[Velocity]
	var spec comp.AccessSpec
	spec.Init(&di, comp.Track(&posCol), comp.Track(&velCol))
	factory := m.CreateFactory(spec)
	factory.Create(1)
	factory.Next()
	id := factory.IDs[0]
	posCol.Slice(&factory.Cursor)[0] = Position{X: 7, Y: 8}

	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()
	// Deliberately reversed relative to registration order at Save time.
	comps := []persist.CompRequest{req[Velocity](&di2), req[Position](&di2)}
	if err := persist.Load(&buf, &di2, &m2.AddressBook, &m2.ArchCatalog, comps); err != nil {
		t.Fatalf("Load with reordered LoadComp: %v", err)
	}

	posID2, _ := di2.ByType(reflect.TypeFor[Position]())
	pos, ok := componentAt[Position](m2, id, posID2.ID)
	if !ok || pos != (Position{X: 7, Y: 8}) {
		t.Errorf("Position after reordered Load = %+v, ok=%v", pos, ok)
	}
}

func TestSaveLoad_MissingLoadComp_ReturnsError(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()
	di.Intern(reflect.TypeFor[Position]())
	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()
	err := persist.Load(&buf, &di2, &m2.AddressBook, &m2.ArchCatalog, nil)
	if err == nil {
		t.Fatal("expected an error when no LoadComp is provided for a required component")
	}
}

func TestSaveLoad_DuplicateLoadComp_Panics(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()
	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	defer func() {
		if recover() == nil {
			t.Error("expected Load to panic on a duplicate LoadComp")
		}
	}()
	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()
	dup := req[Position](&di2)
	_ = persist.Load(&buf, &di2, &m2.AddressBook, &m2.ArchCatalog, []persist.CompRequest{dup, dup})
}

func TestSaveLoad_ExtraLoadComp_RegistersWithoutError(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()
	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()
	// Position never appears in this (empty) save — a "new module" type.
	if err := persist.Load(&buf, &di2, &m2.AddressBook, &m2.ArchCatalog, []persist.CompRequest{req[Position](&di2)}); err != nil {
		t.Fatalf("Load with an unmatched LoadComp: %v", err)
	}
	if _, ok := di2.ByType(reflect.TypeFor[Position]()); !ok {
		t.Error("expected the unmatched component to still be registered")
	}
}

func TestSaveLoad_TruncatedFile_ReturnsError(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()
	var posCol iter.ArrayRef[Position]
	var spec comp.AccessSpec
	spec.Init(&di, comp.Track(&posCol))
	factory := m.CreateFactory(spec)
	factory.Create(1)
	factory.Next()

	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	truncated := buf.Bytes()[:buf.Len()/2]
	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("expected an error, not a panic, for a truncated file: %v", r)
			}
		}()
		err = persist.Load(bytes.NewReader(truncated), &di2, &m2.AddressBook, &m2.ArchCatalog, []persist.CompRequest{req[Position](&di2)})
	}()
	if err == nil {
		t.Fatal("expected an error for a truncated save file")
	}
}

func TestSaveLoad_CorruptedByte_ReturnsError(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	m := newTestManager()
	var posCol iter.ArrayRef[Position]
	var spec comp.AccessSpec
	spec.Init(&di, comp.Track(&posCol))
	factory := m.CreateFactory(spec)
	factory.Create(5)
	for factory.Next() {
		positions := posCol.Slice(&factory.Cursor)
		for i := range positions {
			positions[i] = Position{X: float32(i), Y: float32(i)}
		}
	}

	var buf bytes.Buffer
	if err := persist.Save(&buf, &di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}

	corrupted := append([]byte(nil), buf.Bytes()...)
	mid := len(corrupted) / 2
	corrupted[mid] ^= 0xFF

	var di2 comp.DefIndex
	di2.Init()
	m2 := newTestManager()
	err := persist.Load(bytes.NewReader(corrupted), &di2, &m2.AddressBook, &m2.ArchCatalog, []persist.CompRequest{req[Position](&di2)})
	if err == nil {
		t.Fatal("expected an error for a save file with a corrupted byte")
	}
}
