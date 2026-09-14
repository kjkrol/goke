package persist

import (
	"bytes"
	"reflect"
	"testing"
	"unsafe"
)

// blankFiller declares its alignment filler explicitly; persist must round-trip
// it without tripping over the blank field.
type blankFiller struct {
	A, B uint64
	X, Y bool
	_    [6]byte
}

func TestValue_RoundTripsTypeWithBlankFillerField(t *testing.T) {
	src := blankFiller{A: 7, B: 9, X: true}

	var buf bytes.Buffer
	if err := EncodeValue(&buf, reflect.TypeFor[blankFiller](), unsafe.Pointer(&src)); err != nil {
		t.Fatalf("EncodeValue: %v", err)
	}

	if got, want := buf.Len(), 8+8+1+1; got != want {
		t.Errorf("encoded size = %d, want %d (blank field must not be written)", got, want)
	}

	var dst blankFiller
	if err := DecodeValue(&buf, reflect.TypeFor[blankFiller](), unsafe.Pointer(&dst)); err != nil {
		t.Fatalf("DecodeValue: %v", err)
	}
	if dst != src {
		t.Errorf("round-tripped %+v, want %+v", dst, src)
	}
	if buf.Len() != 0 {
		t.Errorf("%d bytes left unread — encoder and decoder disagree on the blank field", buf.Len())
	}
}
