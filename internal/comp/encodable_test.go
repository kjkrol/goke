package comp_test

import (
	"bytes"
	"log"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/comp"
)

// marshaledType implements encoding.BinaryMarshaler/BinaryUnmarshaler and
// deliberately contains a raw pointer field — ValidateEncodable must treat
// it as opaque and never inspect that field.
type marshaledType struct {
	ptr *int
}

func (m marshaledType) MarshalBinary() ([]byte, error)     { return []byte{1}, nil }
func (m *marshaledType) UnmarshalBinary(data []byte) error { return nil }

type withString struct {
	Name string
	X    float32
}

type withMarshaled struct {
	Tag marshaledType
	Y   int32
}

type nestedPod struct {
	Inner withString
	Count [3]int32
}

type hasPtr struct{ V *int }
type hasSlice struct{ V []int }
type hasMap struct{ V map[string]int }
type hasInterface struct{ V any }
type hasChan struct{ V chan int }
type hasFunc struct{ V func() }
type hasUnsafePointer struct{ V unsafe.Pointer }
type nestedBad struct{ Inner hasPtr }

func TestValidateEncodable_Accepts(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeFor[bool](),
		reflect.TypeFor[int](),
		reflect.TypeFor[uint8](),
		reflect.TypeFor[float64](),
		reflect.TypeFor[string](),
		reflect.TypeFor[position](),
		reflect.TypeFor[[4]float32](),
		reflect.TypeFor[withString](),
		reflect.TypeFor[withMarshaled](),
		reflect.TypeFor[nestedPod](),
		reflect.TypeFor[marshaledType](),
	}
	for _, ty := range types {
		if err := comp.ValidateEncodable(ty); err != nil {
			t.Errorf("expected %s to be encodable, got error: %v", ty, err)
		}
	}
}

func TestValidateEncodable_Rejects(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeFor[hasPtr](),
		reflect.TypeFor[hasSlice](),
		reflect.TypeFor[hasMap](),
		reflect.TypeFor[hasInterface](),
		reflect.TypeFor[hasChan](),
		reflect.TypeFor[hasFunc](),
		reflect.TypeFor[hasUnsafePointer](),
		reflect.TypeFor[nestedBad](),
	}
	for _, ty := range types {
		if err := comp.ValidateEncodable(ty); err == nil {
			t.Errorf("expected %s to be rejected as not encodable", ty)
		}
	}
}

func TestValidateEncodable_NestedFieldPath(t *testing.T) {
	err := comp.ValidateEncodable(reflect.TypeFor[nestedBad]())
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "Inner.V") {
		t.Errorf("expected error to name the nested field path Inner.V, got: %v", err)
	}
}

func TestOffChunkFields(t *testing.T) {
	fields := comp.OffChunkFields(reflect.TypeFor[withMarshaled]())
	if len(fields) != 1 || fields[0] != "Tag" {
		t.Errorf("expected exactly [\"Tag\"], got %v", fields)
	}

	if fields := comp.OffChunkFields(reflect.TypeFor[position]()); len(fields) != 0 {
		t.Errorf("expected no off-chunk fields for a pure POD type, got %v", fields)
	}
}

func TestDefIndex_Intern_LogsOffChunkFieldsOncePerType(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	var c comp.DefIndex
	c.Init()

	c.Intern(reflect.TypeFor[withString]())
	firstCount := strings.Count(buf.String(), "Name")
	if firstCount != 1 {
		t.Errorf("expected exactly one warning after first Intern, got %d in: %s", firstCount, buf.String())
	}

	c.Intern(reflect.TypeFor[withString]()) // idempotent re-registration
	secondCount := strings.Count(buf.String(), "Name")
	if secondCount != firstCount {
		t.Errorf("expected no additional warning on repeat Intern, got %d occurrences", secondCount)
	}
}

func TestDefIndex_Intern_NoLogForPurePOD(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	var c comp.DefIndex
	c.Init()
	c.Intern(reflect.TypeFor[position]())

	if buf.Len() != 0 {
		t.Errorf("expected no warning for a pure POD type, got: %s", buf.String())
	}
}
