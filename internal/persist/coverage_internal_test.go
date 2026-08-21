package persist

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/iter"
)

type covPosition struct{ X, Y float32 }
type covName struct{ Value string }
type covStamp struct{ V int64 }
type covWide struct {
	I16  int16
	U16  uint16
	C64  complex64
	C128 complex128
}

func (s covStamp) MarshalBinary() ([]byte, error) { return fmt.Appendf(nil, "%d", s.V), nil }
func (s *covStamp) UnmarshalBinary(data []byte) error {
	_, err := fmt.Sscanf(string(data), "%d", &s.V)
	return err
}

// buildCoverageWorld returns a DefIndex/Manager spanning three archetypes —
// a fixed-size POD field, a string field, and a BinaryMarshaler field — so
// one saveTo/loadFrom pass touches every encoding path: header, ID pool
// state, component directory, archetype directory, entity IDs, and all
// three value-encoding shapes.
func buildCoverageWorld(t *testing.T) (*comp.DefIndex, *ent.Manager) {
	t.Helper()
	prev := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(prev) })

	var di comp.DefIndex
	di.Init()
	var m ent.Manager
	m.Init(ent.DefaultConfig(), nil)

	var posCol iter.ArrayRef[covPosition]
	var specPos comp.AccessSpec
	specPos.Init(&di, comp.Track(&posCol))
	fPos := m.CreateFactory(specPos)
	fPos.Create(2)
	for fPos.Next() {
		positions := posCol.Slice(&fPos.Cursor)
		for i := range positions {
			positions[i] = covPosition{X: float32(i), Y: float32(i)}
		}
	}

	var nameCol iter.ArrayRef[covName]
	var specName comp.AccessSpec
	specName.Init(&di, comp.Track(&nameCol))
	fName := m.CreateFactory(specName)
	fName.Create(2)
	for fName.Next() {
		names := nameCol.Slice(&fName.Cursor)
		for i := range names {
			names[i] = covName{Value: "entity"}
		}
	}

	var stampCol iter.ArrayRef[covStamp]
	var specStamp comp.AccessSpec
	specStamp.Init(&di, comp.Track(&stampCol))
	fStamp := m.CreateFactory(specStamp)
	fStamp.Create(2)
	for fStamp.Next() {
		stamps := stampCol.Slice(&fStamp.Cursor)
		for i := range stamps {
			stamps[i] = covStamp{V: int64(i)}
		}
	}

	var wideCol iter.ArrayRef[covWide]
	var specWide comp.AccessSpec
	specWide.Init(&di, comp.Track(&wideCol))
	fWide := m.CreateFactory(specWide)
	fWide.Create(2)
	for fWide.Next() {
		values := wideCol.Slice(&fWide.Cursor)
		for i := range values {
			values[i] = covWide{
				I16: int16(i), U16: uint16(i),
				C64: complex(float32(i), float32(-i)), C128: complex(float64(i), float64(-i)),
			}
		}
	}

	return &di, &m
}

// countingWriter never fails — used to count how many Write calls a full,
// successful saveTo makes, so every one of them can be individually failed.
type countingWriter struct{ n int }

func (w *countingWriter) Write(p []byte) (int, error) {
	w.n++
	return len(p), nil
}

// failAfterNWriter succeeds for the first n Write calls, then fails.
type failAfterNWriter struct{ n int }

func (w *failAfterNWriter) Write(p []byte) (int, error) {
	if w.n <= 0 {
		return 0, errors.New("forced write failure")
	}
	w.n--
	return len(p), nil
}

func TestSaveTo_PropagatesEveryWriteError(t *testing.T) {
	di, m := buildCoverageWorld(t)

	var cw countingWriter
	if err := saveTo(&cw, di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("saveTo with a never-failing writer: %v", err)
	}

	for n := range cw.n {
		if err := saveTo(&failAfterNWriter{n: n}, di, &m.AddressBook, &m.ArchCatalog); err == nil {
			t.Errorf("expected saveTo to fail when the writer fails after %d successful writes", n)
		}
	}
}

func covReq[T any](di *comp.DefIndex) CompRequest {
	rt := reflect.TypeFor[T]()
	return CompRequest{
		Name: rt.String(),
		Register: func(wantSize *uint32) error {
			di.Intern(rt)
			return nil
		},
	}
}

func TestLoadFrom_PropagatesEveryReadError(t *testing.T) {
	di, m := buildCoverageWorld(t)

	var buf bytes.Buffer
	if err := saveTo(&buf, di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("saveTo: %v", err)
	}
	valid := buf.Bytes()

	load := func(data []byte) error {
		var di2 comp.DefIndex
		di2.Init()
		var m2 ent.Manager
		m2.Init(ent.DefaultConfig(), nil)
		comps := []CompRequest{covReq[covPosition](&di2), covReq[covName](&di2), covReq[covStamp](&di2), covReq[covWide](&di2)}
		return loadFrom(bytes.NewReader(data), &di2, &m2.AddressBook, &m2.ArchCatalog, comps)
	}

	if err := load(valid); err != nil {
		t.Fatalf("loadFrom with the full, valid payload: %v", err)
	}

	for k := range len(valid) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("loadFrom panicked on a %d-byte prefix instead of returning an error: %v", k, r)
				}
			}()
			if err := load(valid[:k]); err == nil {
				t.Errorf("expected loadFrom to fail on a truncated (%d of %d bytes) payload", k, len(valid))
			}
		}()
	}
}

func TestReadHeader_BadMagic(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("XXXX")
	_ = writeUint32(&buf, FormatVersion)
	if err := readHeader(&buf); err == nil {
		t.Fatal("expected an error for a bad magic")
	}
}

func TestReadHeader_BadVersion(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(Magic)
	_ = writeUint32(&buf, FormatVersion+1)
	if err := readHeader(&buf); err == nil {
		t.Fatal("expected an error for an unsupported format version")
	}
}

func TestReserveBatches_ZeroCount(t *testing.T) {
	if batches := reserveBatches(nil, 0); batches != nil {
		t.Errorf("expected no batches for a zero count, got %v", batches)
	}
}

func TestRegisterComponents_DuplicateName_Panics(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	req := covReq[covPosition](&di)
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic for a duplicate LoadComp registration")
		}
	}()
	_ = registerComponents(nil, []CompRequest{req, req})
}

func TestRegisterComponents_MissingHeader_ReturnsError(t *testing.T) {
	headers := []compHeader{{Name: "an.unknown.Type", Size: 4, Align: 4}}
	if err := registerComponents(headers, nil); err == nil {
		t.Fatal("expected an error when the save file needs a component with no matching LoadComp")
	}
}

func covReqErr[T any]() CompRequest {
	rt := reflect.TypeFor[T]()
	return CompRequest{
		Name: rt.String(),
		Register: func(wantSize *uint32) error {
			return errors.New("forced registration failure")
		},
	}
}

func TestRegisterComponents_MatchedRegisterError_Propagates(t *testing.T) {
	headers := []compHeader{{Name: reflect.TypeFor[covPosition]().String(), Size: 4, Align: 4}}
	if err := registerComponents(headers, []CompRequest{covReqErr[covPosition]()}); err == nil {
		t.Fatal("expected the matched Register's error to propagate")
	}
}

func TestRegisterComponents_UnmatchedRegisterError_Propagates(t *testing.T) {
	if err := registerComponents(nil, []CompRequest{covReqErr[covPosition]()}); err == nil {
		t.Fatal("expected the unmatched (forward-compat) Register's error to propagate")
	}
}

func TestRegisterComponents_UnmatchedForwardCompat(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	req := covReq[covPosition](&di)
	if err := registerComponents(nil, []CompRequest{req}); err != nil {
		t.Fatalf("expected a LoadComp with no matching header to register anyway (forward compatibility): %v", err)
	}
	if di.Count() != 1 {
		t.Fatalf("expected the unmatched component to be registered, got count %d", di.Count())
	}
}

func TestAsBinaryMarshaler_NotAddressable(t *testing.T) {
	if _, ok := asBinaryMarshaler(reflect.ValueOf(covStamp{V: 1})); ok {
		t.Fatal("expected a non-addressable value to never be treated as a BinaryMarshaler")
	}
}

func TestAsBinaryUnmarshaler_NotAddressable(t *testing.T) {
	if _, ok := asBinaryUnmarshaler(reflect.ValueOf(covStamp{V: 1})); ok {
		t.Fatal("expected a non-addressable value to never be treated as a BinaryUnmarshaler")
	}
}

type covFailingMarshaler struct{}

func (covFailingMarshaler) MarshalBinary() ([]byte, error) {
	return nil, errors.New("forced marshal failure")
}
func (*covFailingMarshaler) UnmarshalBinary([]byte) error { return nil }

func TestEncodeValue_PropagatesMarshalBinaryError(t *testing.T) {
	v := covFailingMarshaler{}
	if err := encodeValue(&bytes.Buffer{}, reflect.ValueOf(&v).Elem()); err == nil {
		t.Fatal("expected MarshalBinary's error to propagate")
	}
}

func TestEncodeValue_Array_PropagatesElementError(t *testing.T) {
	arr := [2]uint32{1, 2}
	v := reflect.ValueOf(&arr).Elem()
	if err := encodeValue(&failAfterNWriter{n: 0}, v); err == nil {
		t.Fatal("expected an array element's write error to propagate")
	}
}

func TestDecodeValue_Array_PropagatesElementError(t *testing.T) {
	var arr [2]uint32
	v := reflect.ValueOf(&arr).Elem()
	if err := decodeValue(bytes.NewReader(nil), v); err == nil {
		t.Fatal("expected an array element's read error to propagate")
	}
}

func TestDecodeValue_Bool_PropagatesReadError(t *testing.T) {
	var b bool
	v := reflect.ValueOf(&b).Elem()
	if err := decodeValue(bytes.NewReader(nil), v); err == nil {
		t.Fatal("expected a bool's read error to propagate")
	}
}

func TestSave_PropagatesGzipCloseError(t *testing.T) {
	di, m := buildCoverageWorld(t)
	if err := Save(&failAfterNWriter{n: 0}, di, &m.AddressBook, &m.ArchCatalog); err == nil {
		t.Fatal("expected Save to propagate an underlying write failure")
	}
}

func TestLoad_InvalidGzipStream_ReturnsError(t *testing.T) {
	var di comp.DefIndex
	di.Init()
	var m ent.Manager
	m.Init(ent.DefaultConfig(), nil)
	if err := Load(bytes.NewReader([]byte("not a gzip stream")), &di, &m.AddressBook, &m.ArchCatalog, nil); err == nil {
		t.Fatal("expected Load to reject a non-gzip stream")
	}
}

// freshArchetype builds a single-component archetype (covPosition) inside
// an isolated DefIndex/Catalog, without going through ent.Manager/Factory —
// reserveBatches and loadArchetype only need a Table for a composition,
// never actual entities.
func freshArchetype(t *testing.T) (*comp.DefIndex, *arch.Catalog, arch.ID) {
	t.Helper()
	var di comp.DefIndex
	di.Init()
	def := di.Intern(reflect.TypeFor[covPosition]())
	composition := comp.Composition{}.With(def)

	var catalog arch.Catalog
	catalog.Init(func(*arch.Archetype) {})
	archID := catalog.Upsert(composition)
	return &di, &catalog, archID
}

func TestReserveBatches_MultiChunk(t *testing.T) {
	_, catalog, archID := freshArchetype(t)
	table := &catalog.Archetypes[archID].Table

	const count = 10000
	batches := reserveBatches(table, count)
	if len(batches) < 2 {
		t.Fatalf("expected reserveBatches to span multiple chunks for %d entities, got %d batch(es)", count, len(batches))
	}

	total := 0
	for _, b := range batches {
		if b.ptr == nil {
			t.Error("expected a non-nil chunk pointer for every batch")
		}
		total += b.n
	}
	if total != count {
		t.Errorf("expected batches to sum to %d slots, got %d", count, total)
	}
}

func TestLoadArchetype_UnrecognizedEntity_ReturnsError(t *testing.T) {
	di, catalog, _ := freshArchetype(t)
	def := di.ByID(comp.ID(0))
	ah := archHeader{CompIDs: []uint8{uint8(def.ID)}, EntityCount: 1}

	var book addr.Book
	book.Init(16, 16)
	book.RestorePoolState(0, nil, nil) // nextIndex=0: every id fails IsValid's index<nextIndex check

	var buf bytes.Buffer
	if err := writeUint64(&buf, uint64(uid.UID64(0))); err != nil {
		t.Fatalf("writeUint64: %v", err)
	}

	err := loadArchetype(&buf, di, &book, catalog, ah)
	if err == nil {
		t.Fatal("expected an error for an entity id not recognized by the restored id pool state")
	}
	if !strings.Contains(err.Error(), "not recognized") {
		t.Errorf("expected the error to mention \"not recognized\", got: %v", err)
	}
}

func TestLoad_TrailingData_ReturnsError(t *testing.T) {
	di, m := buildCoverageWorld(t)
	var buf bytes.Buffer
	if err := Save(&buf, di, &m.AddressBook, &m.ArchCatalog); err != nil {
		t.Fatalf("Save: %v", err)
	}
	buf.WriteByte(0)

	var di2 comp.DefIndex
	di2.Init()
	var m2 ent.Manager
	m2.Init(ent.DefaultConfig(), nil)
	comps := []CompRequest{covReq[covPosition](&di2), covReq[covName](&di2), covReq[covStamp](&di2), covReq[covWide](&di2)}
	if err := Load(bytes.NewReader(buf.Bytes()), &di2, &m2.AddressBook, &m2.ArchCatalog, comps); err == nil {
		t.Fatal("expected Load to reject a save file with trailing data after the payload")
	}
}
