package persist_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/persist"
)

func roundTrip[T any](t *testing.T, value T) T {
	t.Helper()
	var buf bytes.Buffer
	if err := persist.EncodeValue(&buf, reflect.TypeFor[T](), unsafe.Pointer(&value)); err != nil {
		t.Fatalf("EncodeValue: %v", err)
	}
	var out T
	if err := persist.DecodeValue(&buf, reflect.TypeFor[T](), unsafe.Pointer(&out)); err != nil {
		t.Fatalf("DecodeValue: %v", err)
	}
	return out
}

func TestValue_RoundTrip_Bool(t *testing.T) {
	for _, v := range []bool{true, false} {
		if got := roundTrip(t, v); got != v {
			t.Errorf("bool %v: got %v", v, got)
		}
	}
}

func TestValue_RoundTrip_Integers(t *testing.T) {
	if got := roundTrip(t, int8(-128)); got != -128 {
		t.Errorf("int8: got %v", got)
	}
	if got := roundTrip(t, int16(-32768)); got != -32768 {
		t.Errorf("int16: got %v", got)
	}
	if got := roundTrip(t, int32(-1<<30)); got != -1<<30 {
		t.Errorf("int32: got %v", got)
	}
	if got := roundTrip(t, int64(-1<<62)); got != -1<<62 {
		t.Errorf("int64: got %v", got)
	}
	if got := roundTrip(t, int(-12345)); got != -12345 {
		t.Errorf("int: got %v", got)
	}
	if got := roundTrip(t, uint8(255)); got != 255 {
		t.Errorf("uint8: got %v", got)
	}
	if got := roundTrip(t, uint16(65535)); got != 65535 {
		t.Errorf("uint16: got %v", got)
	}
	if got := roundTrip(t, uint32(1<<31)); got != 1<<31 {
		t.Errorf("uint32: got %v", got)
	}
	if got := roundTrip(t, uint64(1<<63)); got != 1<<63 {
		t.Errorf("uint64: got %v", got)
	}
	if got := roundTrip(t, uint(999999)); got != 999999 {
		t.Errorf("uint: got %v", got)
	}
}

func TestValue_RoundTrip_Floats(t *testing.T) {
	if got := roundTrip(t, float32(-3.5)); got != -3.5 {
		t.Errorf("float32: got %v", got)
	}
	if got := roundTrip(t, float64(2.71828182845)); got != 2.71828182845 {
		t.Errorf("float64: got %v", got)
	}
}

func TestValue_RoundTrip_Complex(t *testing.T) {
	if got := roundTrip(t, complex64(1.5-2.5i)); got != complex64(1.5-2.5i) {
		t.Errorf("complex64: got %v", got)
	}
	if got := roundTrip(t, complex128(3.14+2.71i)); got != complex128(3.14+2.71i) {
		t.Errorf("complex128: got %v", got)
	}
}

func TestValue_RoundTrip_String(t *testing.T) {
	for _, s := range []string{"", "hello", "zółw zółc plź niebieski", string(make([]byte, 5000))} {
		if got := roundTrip(t, s); got != s {
			t.Errorf("string %q: got %q", s, got)
		}
	}
}

type roundTripStruct struct {
	Name  string
	X, Y  float32
	Count int32
	Tags  [3]uint8
}

func TestValue_RoundTrip_NestedStruct(t *testing.T) {
	v := roundTripStruct{Name: "order-1", X: 1.5, Y: -2.5, Count: 7, Tags: [3]uint8{1, 2, 3}}
	got := roundTrip(t, v)
	if got != v {
		t.Errorf("struct: got %+v, want %+v", got, v)
	}
}

// marshaledPoint implements encoding.BinaryMarshaler/BinaryUnmarshaler and
// deliberately carries an unexported pointer field — a correct
// implementation never inspects it directly, only through the interface.
type marshaledPoint struct {
	x, y int32
}

func (p marshaledPoint) MarshalBinary() ([]byte, error) {
	return fmt.Appendf(nil, "%d,%d", p.x, p.y), nil
}

func (p *marshaledPoint) UnmarshalBinary(data []byte) error {
	_, err := fmt.Sscanf(string(data), "%d,%d", &p.x, &p.y)
	return err
}

func TestValue_RoundTrip_BinaryMarshaler(t *testing.T) {
	v := marshaledPoint{x: 10, y: -20}
	got := roundTrip(t, v)
	if got != v {
		t.Errorf("marshaledPoint: got %+v, want %+v", got, v)
	}
}

type withMarshaledField struct {
	Tag   marshaledPoint
	Extra int64
}

func TestValue_RoundTrip_StructWithMarshaledField(t *testing.T) {
	v := withMarshaledField{Tag: marshaledPoint{x: 1, y: 2}, Extra: 42}
	got := roundTrip(t, v)
	if got != v {
		t.Errorf("withMarshaledField: got %+v, want %+v", got, v)
	}
}

// EncodeValue/DecodeValue trust comp.ValidateEncodable to have already
// rejected an unencodable kind at RegComp — reaching the encoder/decoder
// with one anyway (a defensive, "should never happen" scenario) panics
// rather than silently producing garbage.
func TestValue_PanicsOnUnencodableKind(t *testing.T) {
	type hasSlice struct{ V []int }
	var v hasSlice

	t.Run("encode", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected EncodeValue to panic on an unencodable kind")
			}
		}()
		var buf bytes.Buffer
		_ = persist.EncodeValue(&buf, reflect.TypeFor[hasSlice](), unsafe.Pointer(&v))
	})

	t.Run("decode", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected DecodeValue to panic on an unencodable kind")
			}
		}()
		r := bytes.NewReader(nil)
		_ = persist.DecodeValue(r, reflect.TypeFor[hasSlice](), unsafe.Pointer(&v))
	})
}
