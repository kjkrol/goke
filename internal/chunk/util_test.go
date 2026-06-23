package chunk

import (
	"testing"
	"unsafe"
)

func TestCopyMemory(t *testing.T) {
	src := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	dst := make([]byte, len(src))

	CopyMemory(unsafe.Pointer(&dst[0]), unsafe.Pointer(&src[0]), uintptr(len(src)))

	for i := range src {
		if dst[i] != src[i] {
			t.Errorf("byte %d: expected %d, got %d", i, src[i], dst[i])
		}
	}
}

func TestZeroMemory(t *testing.T) {
	buf := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	ZeroMemory(unsafe.Pointer(&buf[0]), uintptr(len(buf)))

	for i, b := range buf {
		if b != 0 {
			t.Errorf("byte %d: expected 0, got %d", i, b)
		}
	}
}
